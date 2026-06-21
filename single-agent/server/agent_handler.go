package server

// https://github.com/achetronic/adk-utils-go.git
import (
	"context"
	"fmt"
	"log"
	"net/http"

	"single-agent/agents"
	config "single-agent/config"
	mymcp "single-agent/mcp"

	"single-agent/observability"
	// "single-agent/provider/gemini"

	"single-agent/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/memory"
	"google.golang.org/adk/model"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool/toolconfirmation"
	"google.golang.org/genai"
)

const (
	appName = "ecommerce"
	userID  = "demo_user"
)

type ChatRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

type ChatResponse struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`

	RequiresConfirmation bool        `json:"requires_confirmation"`
	ConfirmationID       string      `json:"confirmation_id"`
	Hint                 string      `json:"hint"`
	Payload              interface{} `json:"payload"`
}

type ConfirmRequest struct {
	SessionID      string      `json:"session_id"`
	Confirmed      bool        `json:"confirmed"`
	ConfirmationID string      `json:"confirmation_id"`
	Hint           string      `json:"hint"`
	Payload        interface{} `json:"payload"`
}

type AgentServer struct {
	Runner         *runner.Runner
	SessionService session.Service
	Config         *config.AppConfig
	Telemetry      *observability.Telemetry
}

type contextKey string

const contextKeyAuthToken contextKey = "auth_token"

// headerTransport là một RoundTripper tùy chỉnh để chèn thêm header vào mọi request
// Agent (Start Span) -> RoundTrip (Inject) -> [Network] -> MCP Server (Extract) -> Tool Handler (Inject) -> [Network] -> Backend (Extract)
type headerTransport struct {
	base   http.RoundTripper
	header string
	value  string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone request để không làm ảnh hưởng đến bản gốc
	newReq := req.Clone(req.Context())

	// Check if token is propagated in context
	if tokenVal := req.Context().Value(contextKeyAuthToken); tokenVal != nil {
		if tokenStr, ok := tokenVal.(string); ok && tokenStr != "" {
			newReq.Header.Set(t.header, tokenStr)
		} else {
			newReq.Header.Set(t.header, t.value)
		}
	} else {
		newReq.Header.Set(t.header, t.value)
	}

	propagator := otel.GetTextMapPropagator()
	propagator.Inject(req.Context(), propagation.HeaderCarrier(newReq.Header))
	return t.base.RoundTrip(newReq)
}

func runAgent(ctx context.Context, runnr *runner.Runner, sessionID string, input string) string {
	userMsg := genai.NewContentFromText(input, genai.RoleUser)

	var responseText string
	for event, err := range runnr.Run(ctx, userID, sessionID, userMsg, agent.RunConfig{}) {
		if err != nil {
			log.Printf("Error: %v", err)
			break
		}
		if event.ErrorCode != "" {
			log.Printf("Event error: %s - %s", event.ErrorCode, event.ErrorMessage)
			break
		}
		if event.Content != nil && len(event.Content.Parts) > 0 {
			responseText += event.Content.Parts[0].Text
		}
	}

	return responseText
}

func NewAgentServer(ctx context.Context, appCfg *config.AppConfig, telemetry *observability.Telemetry, llm model.LLM) (*AgentServer, error) {

	// 2. Init Shared Resources
	mcpToken := utils.GetEnvString("AUTH_TOKEN", "")

	// Shared MCP Transport
	transport := &mcp.SSEClientTransport{
		Endpoint: appCfg.McpServer,
		HTTPClient: &http.Client{
			Transport: &headerTransport{
				base:   http.DefaultTransport,
				header: "Authorization",
				value:  "Bearer " + mcpToken,
			},
		},
	}

	// 3. Initialize Single Agent
	agentCfg, ok := appCfg.Agents["ecommerce_agent"]
	if !ok {
		return nil, fmt.Errorf("ecommerce_agent config not found")
	}

	// Initialize MCP tools for this agent
	mcpToolset, err := mymcp.NewMCPTool(transport, agentCfg.AllowedTools, agentCfg.ApprovedTools)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP tools for ecommerce_agent: %w", err)
	}

	// Create the single Agent instance
	targetAgent, err := agents.NewSingleAgent(ctx, &agentCfg, &appCfg.Prompts, llm, mcpToolset)
	if err != nil {
		return nil, fmt.Errorf("failed to create ecommerce agent: %w", err)
	}
	log.Printf("Single agent initialized: %s", agentCfg.Name)

	// 5. Create session and memory
	sessionService := session.InMemoryService()
	// sessionService, err := mySess.NewRedisService(mySess.GetConfigRedis())
	memoryService := memory.InMemoryService()
	// memoryService, err := mySvc.NewPostgresMemoryService(ctx, mySvc.GetConfigPGMem())

	// if err != nil {
	// 	log.Printf("Failed to init session service for agent: %s", err.Error())
	// 	return nil, fmt.Errorf("Failed to init session service for agent: %s", err.Error())
	// }

	runr, err := runner.New(runner.Config{
		AppName:        appName,
		Agent:          targetAgent,
		SessionService: sessionService,
		MemoryService:  memoryService,
	})
	if err != nil {
		log.Printf("Failed to create runner: %v", err)
		return nil, fmt.Errorf("failed to create runner: %w", err)
	}

	return &AgentServer{Runner: runr, SessionService: sessionService, Config: appCfg, Telemetry: telemetry}, nil
}

func (s *AgentServer) HandlerChat(c *gin.Context) {
	var r ChatRequest
	if err := c.ShouldBindBodyWithJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Invalid request": err.Error()})
		return
	}

	if r.SessionID == "" {
		sessionID := uuid.NewString()
		_, err := s.SessionService.Create(c.Request.Context(), &session.CreateRequest{
			UserID:    userID,
			SessionID: sessionID,
			AppName:   appName,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Failed to create session ": err.Error()})
			log.Printf("Failed to create session: %v", err)
			return
		}
		r.SessionID = sessionID
	}

	userMsg := genai.NewContentFromText(r.Message, genai.RoleUser)

	ctx := c.Request.Context()
	if authToken := c.GetHeader("Authorization"); authToken != "" {
		ctx = context.WithValue(ctx, contextKeyAuthToken, authToken)
	}

	ctxOtel, span := s.Telemetry.Tracer.Start(ctx, "agent.request")
	defer span.End()

	// Turn 1: chạy bình thường, capture confirmation event
	var pendingConfirmations map[string]toolconfirmation.ToolConfirmation
	var confirmationCallID string

	finalResponse := ""

	for event, err := range s.Runner.Run(ctxOtel, userID, r.SessionID, userMsg, agent.RunConfig{}) {
		if err != nil {
			log.Printf("Run error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"Failed to run agent ": err.Error()})
		}

		// In text bình thường
		if event.Content != nil && len(event.Content.Parts) > 0 {
			if event.Content.Parts[0].Text != "" {
				finalResponse += event.Content.Parts[0].Text
			}
			for _, part := range event.Content.Parts {
				// Nếu thấy Agent gọi hàm adk_request_confirmation
				if part.FunctionCall != nil && part.FunctionCall.Name == "adk_request_confirmation" {
					confirmationCallID = part.FunctionCall.ID // LẤY CÁI ID NÀY!
					// log.Printf("[DEBUG] Thấy hàm duyệt ảo! ID: %s", confirmationCallID)
				}
			}
		}
		// Capture confirmation đang chờ
		if event.Actions.RequestedToolConfirmations != nil {
			pendingConfirmations = event.Actions.RequestedToolConfirmations
		}
	}
	fmt.Println("Pending confirm:", pendingConfirmations)

	// Turn 2: nếu có confirmation đang chờ → gửi approve response
	if len(pendingConfirmations) > 0 {
		for callID, conf := range pendingConfirmations {
			// log.Printf("[DEBUG] Approving: callID=%s, hint=%s", callID, conf.Hint)

			if confirmationCallID != "" && callID != "" {

				// Dùng confirmationCallID thay vì cái ID trong map RequestedToolConfirmations
				c.JSON(http.StatusOK, ChatResponse{
					SessionID:            r.SessionID,
					Message:              finalResponse,
					RequiresConfirmation: true,
					ConfirmationID:       confirmationCallID, // Trả cái ID ảo này về cho Client
					Hint:                 conf.Hint,          // Lấy hint từ map
				})
				return
			}
		}
	}
	c.JSON(http.StatusOK, ChatResponse{
		SessionID: r.SessionID,
		Message:   finalResponse,
	})
}

func (s *AgentServer) HandlerConfirm(c *gin.Context) {
	var r ConfirmRequest
	if err := c.ShouldBindBodyWithJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Invalid request": err.Error()})
		return
	}

	ctx := c.Request.Context()
	if authToken := c.GetHeader("Authorization"); authToken != "" {
		ctx = context.WithValue(ctx, contextKeyAuthToken, authToken)
	}

	ctxOtel, span := s.Telemetry.Tracer.Start(ctx, "agent.confirm")
	defer span.End()
	confirmationCallID := r.ConfirmationID
	var parts []*genai.Part

	if confirmationCallID != "" {
		// Dùng confirmationCallID thay vì cái ID trong map RequestedToolConfirmations
		parts = append(parts, &genai.Part{
			FunctionResponse: &genai.FunctionResponse{
				ID:   confirmationCallID, // PHẢI DÙNG ID NÀY
				Name: "adk_request_confirmation",
				Response: map[string]any{
					"confirmed": true,
					"hint":      r.Hint,
					"payload":   r.Payload,
				},
			},
		})
	}

	approvalMsg := &genai.Content{
		Role:  string(genai.RoleUser),
		Parts: parts,
	}
	finalResponse := ""
	for event, err := range s.Runner.Run(ctxOtel, userID, r.SessionID, approvalMsg, agent.RunConfig{}) {
		if err != nil {
			log.Printf("Resume error: %v", err)
			// Trả lỗi về để mình biết chính xác tại sao nó không chạy
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return // THOÁT NGAY
		}

		if event.Content != nil {
			for _, part := range event.Content.Parts {
				if part.Text != "" {
					finalResponse += part.Text
				}
			}
		}
	}

	// Trả kết quả cuối cùng
	c.JSON(http.StatusOK, ChatResponse{
		SessionID: r.SessionID,
		Message:   finalResponse,
	})

}
