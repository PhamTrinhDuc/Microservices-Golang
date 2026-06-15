package main

import (
	"context"
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
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	defaultPort      = "8081"
	defaultDBHost    = "localhost"
	defaultDBPort    = 5433
	defaultRedisHost = "localhost"
	defaultRedisPort = 6379
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
		Port: utils.GetEnvString("PORT", defaultPort),
		Database: database.DBConfig{
			Host:     utils.GetEnvString("DB_HOST", defaultDBHost),
			Port:     utils.GetEnvInt("DB_PORT", defaultDBPort),
			User:     utils.GetEnvString("DB_USER", "mcp_user"),
			Password: utils.GetEnvString("DB_PASSWORD", "mcp_password"),
			DBName:   utils.GetEnvString("DB_NAME", "salon_chain"),
			SSLMode:  utils.GetEnvString("DB_SSLMODE", "disable"),
			MaxConns: int32(utils.GetEnvInt("DB_MAX_CONNS", 25)),
			MinConns: int32(utils.GetEnvInt("DB_MIN_CONNS", 5)),
		},
		Redis: redis.RedisConfig{
			Host:     utils.GetEnvString("REDIS_HOST", defaultRedisHost),
			Port:     utils.GetEnvInt("REDIS_PORT", defaultRedisPort),
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
		BackendURL: utils.GetEnvString("BACKEND_URL", "http://localhost:8080/api/v1"),
	}
}

func main() {
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
		authMid.Handler(),
		rateLimiter.Handler(),
	)
	{
		// Official SDK SSE transport needs wildcards for session handling
		mcpGroup.Any("/*path", gin.WrapH(mcpHandler))
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
