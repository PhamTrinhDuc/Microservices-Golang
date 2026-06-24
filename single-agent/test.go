package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"single-agent/config"
	"single-agent/observability"
	"single-agent/plugins/langfuse"
	"single-agent/provider/openai"
	server "single-agent/server"
	"single-agent/utils"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/otel"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

func main() {
	// Load godotenv first
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: failed to load .env file")
	}

	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		log.Printf("[OTel Error Handler] %v", err)
	}))

	pubKey := os.Getenv("LANGFUSE_PUBLIC_KEY")
	secKey := os.Getenv("LANGFUSE_SECRET_KEY")
	fmt.Printf("LANGFUSE_PUBLIC_KEY: %q (len=%d)\n", pubKey, len(pubKey))
	fmt.Printf("LANGFUSE_SECRET_KEY: %q (len=%d)\n", secKey, len(secKey))

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
		APIKey:  utils.GetEnvString("GROQ_API_KEY", ""),
		BaseURL: utils.GetEnvString("BASE_URL_GROQ", "https://api.groq.com/openai/v1"),
		// ModelName: utils.GetEnvString("GROQ_LLM", "qwen/qwen3.6-27b"),
		ModelName: utils.GetEnvString("GROQ_LLM", "qwen/qwen3-32b"),
	})
	appCfg, err := config.LoadAppConfig("./config.yaml")
	if err != nil {
		log.Fatalf("Failed to load app config: %v", err)
	}

	pluginCfg, langfuseShutdown, err := langfuse.Setup(&langfuse.Config{
		PublicKey:   pubKey,
		SecretKey:   secKey,
		Host:        "https://cloud.langfuse.com", // or self-hosted URL
		Environment: "production",
		ServiceName: "my-agent",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if langfuseShutdown != nil {
			fmt.Println("Triggering Langfuse telemetry shutdown...")
			if err := langfuseShutdown(context.Background()); err != nil {
				log.Printf("Langfuse shutdown err: %v", err)
			}
		}
	}()

	agents, err := server.NewAgentServer(ctx, appCfg, telemetry, &pluginCfg, llm)
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

	userMsg := genai.NewContentFromText("Bên bạn có đồng hồ thông minh < 3 triệu mà có thể GPS để chạy bộ không?", genai.RoleUser)

	toolMapping := map[string]string{
		"list_categories":         "Tìm kiếm danh mục sản phẩm phù hợp...",
		"get_specs_by_category":   "Kiểm tra thông số kỹ thuật (Hỗ trợ GPS, Chạy bộ)...",
		"list_products":           "Tìm kiếm các sản phẩm trong khoảng giá yêu cầu...",
		"get_product_by_id":       "Kiểm tra chi tiết thông tin và tồn kho sản phẩm...",
		"hybrid_search_documents": "Tìm kiếm tài liệu chính sách mua sắm/bảo hành...",
		"list_brands":             "Kiểm tra danh sách các thương hiệu...",
	}

	fmt.Println("\n--- Bắt đầu hành trình Agent ---")
	for event, err := range agents.Runner.Run(ctx, "demo_user", sessionID, userMsg, agent.RunConfig{}) {
		if err != nil {
			log.Fatalf("\nRun error: %v", err)
		}
		if event.Content != nil {
			for _, part := range event.Content.Parts {
				if part.FunctionCall != nil {
					toolName := part.FunctionCall.Name
					friendlyName, ok := toolMapping[toolName]
					if !ok {
						friendlyName = fmt.Sprintf("Đang xử lý bước %s...", toolName)
					}
					fmt.Printf("\n[⚙️ Bước đi] %s (Công cụ: %s)\n", friendlyName, toolName)
				}
				if part.FunctionResponse != nil {
					fmt.Printf("[✓ Đã hoàn thành] Nhận dữ liệu kết quả từ công cụ %s\n", part.FunctionResponse.Name)
				}
				if part.Text != "" {
					fmt.Print(part.Text)
				}
			}
		}
	}
	fmt.Println("\n--- Kết thúc hành trình Agent ---")
}
