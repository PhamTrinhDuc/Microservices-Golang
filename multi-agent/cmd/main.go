package main

import (
	"context"
	"log"
	"multi-agent/observability"
	"multi-agent/server"
	"multi-agent/utils"

	"github.com/gin-gonic/gin"
)

func loadConfig() observability.Config {
	return observability.Config{
		ServiceName:    utils.GetEnvString("OTEL_SERVICE", "agent-server"),
		ServiceVersion: utils.GetEnvString("OTEL_VERSION", "1.0.0"),
		Environment:    utils.GetEnvString("ENVIRONMENT", "development"),
		OTLPEndpoint:   utils.GetEnvString("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"),
		SamplingRate:   utils.GetEnvFloat("OTEL_TRACES_SAMPLER_ARG", 1.0),
		EnableTracing:  utils.GetEnvBool("OTEL_ENABLE_TRACING", true),
		EnableMetrics:  utils.GetEnvBool("OTEL_ENABLE_METRICS", true),
	}
}

func main() {
	ctx := context.Background()

	// 1. Init telemetry
	// cfgTelemetry := loadConfig()
	// telemetry, err := observability.NewTelemetry(ctx, cfgTelemetry)
	// if err != nil {
	// 	log.Printf("Failed to initialize telemetry: %v", err)
	// }

	// // tạo context mới để shutdown thay vì shutdown trực tiếp để tránh cancel context chính của app
	// defer func() {
	// 	shudownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 	defer cancel()
	// 	if err := telemetry.Shutdown(shudownCtx); err != nil {
	// 		log.Printf("Error shutting down telemetry: %v", err)
	// 	}
	// }()
	// log.Println("OpenTelemetry initialized successfully")

	// 2. Khởi tạo Agent Server (Nó sẽ tự lo liệu từ Config, LLM đến MCP)
	agentServer, err := server.NewAgentsServer(ctx, "../config.yaml")
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
	log.Println("Agent API Server is running on :8089")
	if err := r.Run(":8089"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
