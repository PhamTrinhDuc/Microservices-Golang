package main

import (
	"context"
	"fmt"
	"log"
	"multi-agent/config"
	memory "multi-agent/memory/postgres"
	session "multi-agent/memory/redis"
	"multi-agent/observability"
	anthropicprd "multi-agent/provider/anthropic"
	openaiprd "multi-agent/provider/openai"
	"multi-agent/server"
	"multi-agent/utils"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)

type Config struct {
	Observability observability.Config
	Memory        memory.PostgresMemoryServiceConfig
	Session       session.RedisConfig
	ConfigPath    string
	Provider      string
	Port          string
}

func NewProvider(provider string) (model.LLM, error) {
	switch provider {
	case "openai":
		return openaiprd.New(openaiprd.Config{
			APIKey:    utils.GetEnvString("OPENAI_API_KEY", ""),
			ModelName: utils.GetEnvString("OPENAI_LLM", "gpt-4o-mini"),
		}), nil
	case "gemini":
		return gemini.NewModel(
			context.Background(),
			utils.GetEnvString("GEMINI_LLM", ""),
			&genai.ClientConfig{
				APIKey: utils.GetEnvString("GEMINI_API_KEY", ""),
			})
	case "claude":
		return anthropicprd.New(anthropicprd.Config{
			APIKey:    utils.GetEnvString("CLAUDE_API_KEY", ""),
			ModelName: utils.GetEnvString("CLAUDE_LLM", "claude-3-5-sonnet"),
		}), nil
	case "groq":
		return openaiprd.New(
			openaiprd.Config{
				APIKey:    utils.GetEnvString("GROQ_API_KEY", ""),
				BaseURL:   utils.GetEnvString("BASE_URL_GROQ", "https://api.groq.com/openai/v1"),
				ModelName: utils.GetEnvString("GROQ_LLM", "qwen/qwen3.6-27b"),
			}), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}

func loadConfig() Config {
	return Config{
		Observability: observability.Config{
			ServiceName:    utils.GetEnvString("OTEL_SERVICE", "agent-server"),
			ServiceVersion: utils.GetEnvString("OTEL_VERSION", "1.0.0"),
			Environment:    utils.GetEnvString("ENVIRONMENT", "development"),
			OTLPEndpoint:   utils.GetEnvString("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"),
			SamplingRate:   utils.GetEnvFloat("OTEL_TRACES_SAMPLER_ARG", 1.0),
			EnableTracing:  utils.GetEnvBool("OTEL_ENABLE_TRACING", true),
			EnableMetrics:  utils.GetEnvBool("OTEL_ENABLE_METRICS", true),
		},
		Memory: memory.PostgresMemoryServiceConfig{
			ConnString: utils.GetEnvString("POSTGRES_URL", "postgres://jiyuu_user:jiyuu_password@localhost:5433/ecommerce_db"),
			EmbeddingModel: memory.NewOpenAICompatibleEmbedding(
				memory.OpenAICompatibleEmbeddingConfig{
					// BaseURL:   utils.GetEnvString("BASE_URL_OLLAMA", "http://localhost:11434/v1"),
					// Model:     utils.GetEnvString("EMBEDDING_MODEL", "qwen3-embedding:0.6b"),
					BaseURL:   utils.GetEnvString("OPENAI_BASE_URL", "https://api.openai.com/v1"),
					Model:     utils.GetEnvString("OPENAI_EMBEDDING_MODEL", "text-embedding-3-large"),
					Dimension: utils.GetEnvInt("OPENAI_EMBEDDING_DIM", 675),
					APIKey:    utils.GetEnvString("OPENAI_API_KEY", ""),
				},
			),
			TopKBM25:     50,
			TopKVector:   50,
			TopKHybrid:   5,
			WeightBM25:   0.5,
			WeightVector: 0.5,
		},
		Session: session.RedisConfig{
			Host:         utils.GetEnvString("REDIS_HOST", "localhost"),
			Port:         utils.GetEnvInt("REDIS_PORT", 6379),
			Username:     utils.GetEnvString("REDIS_USERNAME", "jiyuu"),
			Password:     utils.GetEnvString("REDIS_PASSWORD", "a2amcpgo"),
			AppStateTTL:  1 * time.Second,
			UserStateTTL: 1 * time.Second,
			TTL:          15 * time.Second,
		},

		ConfigPath: "../config.yaml",
		Provider:   utils.GetEnvString("PROVIDER", "groq"),
		Port:       utils.GetEnvString("AGENT_SERVER_PORT", "8000"),
	}
}

func main() {
	ctx := context.Background()

	// 1. Init telemetry
	cfg := loadConfig()
	telemetry, err := observability.NewTelemetry(ctx, cfg.Observability)
	if err != nil {
		log.Printf("Failed to initialize telemetry: %v", err)
	}

	// tạo context mới để shutdown thay vì shutdown trực tiếp để tránh cancel context chính của app
	defer func() {
		shudownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := telemetry.Shutdown(shudownCtx); err != nil {
			log.Printf("Error shutting down telemetry: %v", err)
		}
	}()
	log.Println("OpenTelemetry initialized successfully")

	// 2. Khởi tạo Agent Server (Nó sẽ tự lo liệu từ Config, LLM đến MCP)
	appCfg, err := config.LoadAppConfig(cfg.ConfigPath)
	if err != nil {
		log.Fatalf("Failed to load app config: %v", err)
	}

	llm, err := NewProvider(cfg.Provider)
	if err != nil {
		log.Fatalf("Failed to create provider: %v", err)
	}

	agentServer, err := server.NewAgentServer(ctx, appCfg, telemetry, llm)
	if err != nil {
		log.Fatalf("Failed to initialize Agent Server: %v", err)
	}

	// 3. Thiết lập Gin
	r := gin.Default()

	// 4. Đăng ký các Routes
	api := r.Group("/api")
	{
		api.POST("/chat", agentServer.HandlerChat)
		api.POST("/chat/confirm", agentServer.HandlerConfirm)
	}

	// 5. Chạy Server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Agent server starting on :%s", cfg.Port)
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
