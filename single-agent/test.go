package main

import (
	"context"
	"fmt"
	"log"

	"single-agent/config"
	"single-agent/observability"
	"single-agent/provider/openai"
	server "single-agent/server"
	"single-agent/utils"

	"github.com/google/uuid"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

func main() {
	ctx := context.Background()
	telemetry, _ := observability.NewTelemetry(ctx, observability.Config{
		ServiceName:    utils.GetEnvString("OTEL_SERVICE", "mcp-server"),
		ServiceVersion: utils.GetEnvString("OTEL_VERSION", "1.0.0"),
		Environment:    utils.GetEnvString("ENVIRONMENT", "development"),
		OTLPEndpoint:   utils.GetEnvString("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4318"),
		SamplingRate:   utils.GetEnvFloat("OTEL_TRACES_SAMPLER_ARG", 1.0),
		EnableTracing:  utils.GetEnvBool("OTEL_ENABLE_TRACING", true),
		EnableMetrics:  utils.GetEnvBool("OTEL_ENABLE_METRICS", true),
	})

	llm := openai.New(openai.Config{
		APIKey:    utils.GetEnvString("GROQ_API_KEY", ""),
		BaseURL:   utils.GetEnvString("BASE_URL_GROQ", "https://api.groq.com/openai/v1"),
		ModelName: utils.GetEnvString("GROQ_LLM", "qwen/qwen3.6-27b"),
		// ModelName: utils.GetEnvString("GROQ_LLM", "qwen/qwen3-32b"),
	})
	appCfg, err := config.LoadAppConfig("./config.yaml")
	if err != nil {
		log.Fatalf("Failed to load app config: %v", err)
	}

	agents, err := server.NewAgentServer(ctx, appCfg, telemetry, llm)
	if err != nil {
		log.Fatalf("Failed to initialize AgentsServer: %v", err)
	}

	sessionID := uuid.New().String()
	_, err = agents.SessionService.Create(ctx, &session.CreateRequest{
		UserID:    "demo_user",
		SessionID: sessionID,
		AppName:   "ecommerce",
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	userMsg := genai.NewContentFromText("Tôi đang cần 1 con màn hình rời 27inch, 2K mà giá loanh quanh 4 triệu?", genai.RoleUser)

	fmt.Printf("Agent: ")
	for event, err := range agents.Runner.Run(ctx, "demo_user", sessionID, userMsg, agent.RunConfig{}) {
		if err != nil {
			log.Fatalf("\nRun error: %v", err)
		}
		if event.Content != nil && len(event.Content.Parts) > 0 {
			fmt.Print(event.Content.Parts[0].Text)
		}
	}
}
