package main

import (
	"context"
	"fmt"
	"log"
	"mcp-server/internal/database"
	"mcp-server/internal/middleware"
	"mcp-server/internal/observability"
	"mcp-server/internal/redis"
	"mcp-server/internal/server"
	"mcp-server/internal/utils"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Config struct {
	Port          string
	Database      database.DBConfig
	Redis         redis.RedisConfig
	Telemetry     observability.Config
	EnableTracing bool
	EnableMetrics bool
	BackendURL    string
}

func loadConfig() Config {
	return Config{
		Port: utils.GetEnvString("MCP_SERVER_PORT", "8081"),
		Database: database.DBConfig{
			Host:     utils.GetEnvString("DB_HOST", "localhost"),
			Port:     utils.GetEnvInt("DB_PORT", 5433),
			User:     utils.GetEnvString("DB_USER", "jiyuu_user"),
			Password: utils.GetEnvString("DB_PASSWORD", "jiyuu_password"),
			DBName:   utils.GetEnvString("DB_NAME", "ecommerce_db"),
			SSLMode:  utils.GetEnvString("DB_SSLMODE", "disable"),
			MaxConns: int32(utils.GetEnvInt("DB_MAX_CONNS", 25)),
			MinConns: int32(utils.GetEnvInt("DB_MIN_CONNS", 5)),
		},
		Redis: redis.RedisConfig{
			Host:     utils.GetEnvString("REDIS_HOST", "localhost"),
			Port:     utils.GetEnvInt("REDIS_PORT", 6379),
			Username: utils.GetEnvString("REDIS_USERNAME", "jiyuu"),
			Password: utils.GetEnvString("REDIS_PASSWORD", "a2amcpgo"),
			DB:       utils.GetEnvInt("REDIS_DB", 0),
			PoolSize: utils.GetEnvInt("REDIS_POOL_SIZE", 10),
			MinCons:  utils.GetEnvInt("REDIS_MIN_CONNS", 2),
		},
		Telemetry: observability.Config{
			ServiceName:    utils.GetEnvString("OTEL_SERVICE", "mcp-server"),
			ServiceVersion: utils.GetEnvString("OTEL_VERSION", "1.0.0"),
			Environment:    utils.GetEnvString("ENVIRONMENT", "development"),
			OTLPEndpoint:   utils.GetEnvString("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"),
			SamplingRate:   utils.GetEnvFloat("OTEL_TRACES_SAMPLER_ARG", 1.0),
			EnableTracing:  utils.GetEnvBool("OTEL_ENABLE_TRACING", true),
			EnableMetrics:  utils.GetEnvBool("OTEL_ENABLE_METRICS", true),
		},
		BackendURL: utils.GetEnvString("BACKEND_URL", fmt.Sprintf("http://localhost:%s/api/v1", utils.GetEnvString("BACKEND_PORT", "8082"))),
	}
}

func main() {
	// Load godotenv first
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: failed to load .env file, reading system environment variables")
	}

	// 1. Init context and config
	ctx := context.Background()
	cfg := loadConfig()

	// 2. Init Services
	// a. Init Database
	db, err := database.NewDB(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("DB fail: %v", err)
	}
	defer db.Close()
	// b. Init Redis
	redis, err := redis.NewRedis(ctx, cfg.Redis)
	if err != nil {
		log.Fatalf("Redis fail: %v", err)
	}
	defer redis.Close()

	// c. Init Telemetry
	telemetry, err := observability.NewTelemetry(ctx, cfg.Telemetry)
	if err != nil {
		log.Fatalf("Failed to initialize telemetry: %v", err)
	}

	// tạo context mới để shutdown thay vì shutdown trực tiếp để tránh cancel context chính của app
	defer func() {
		shudownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := telemetry.Shutdown(shudownCtx); err != nil {
			log.Fatalf("Error shutting down telemetry: %v", err)
		}
	}()
	log.Println("OpenTelemetry initialized successfully")

	// d. Init middleware
	authMid := middleware.NewAuthMiddleware()
	rateLimiter := middleware.NewFixedWindowLimiter(redis.Client, 100, time.Minute)
	tracingMiddleware := middleware.NewTracingMiddleware(telemetry)

	r := gin.Default()

	r.Use(tracingMiddleware.Handler())

	// 3. Health & Metrics
	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Welcome to A2A MCP Server",
			"version": "1.0.0",
			"status":  "running",
		})
	})
	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	if cfg.EnableMetrics {
		r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	// 4. MCP Endpoint with Gin
	mcpHandler := server.NewSSEHandler(db, telemetry, cfg.BackendURL)

	mcpGroup := r.Group("/mcp")
	mcpGroup.Use(
		authMid.OptionalHandler(), // Dùng OptionalHandler để cho phép kết nối SSE và ListTools không cần Token
		rateLimiter.Handler(),
	)
	{
		// Đọc thông tin từ Gin context chuyển vào HTTP Request context trước khi chạy MCP Handler
		mcpGroup.Any("/*path", func(c *gin.Context) {
			ctx := c.Request.Context()
			for _, key := range []string{"user_id", "email", "role"} {
				if val, exists := c.Get(key); exists {
					ctx = context.WithValue(ctx, key, val)
				}
			}
			c.Request = c.Request.WithContext(ctx)
			mcpHandler.ServeHTTP(c.Writer, c.Request)
		})
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("MCP Server starting on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
