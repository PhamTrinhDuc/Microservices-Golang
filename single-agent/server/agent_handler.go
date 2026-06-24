package server

// https://github.com/achetronic/adk-utils-go.git
import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"log"
	"log/slog"
	"net/http"

	"single-agent/agents"
	config "single-agent/config"
	mymcp "single-agent/mcp"

	"single-agent/observability"
	// "single-agent/provider/gemini"

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
	appName      = "ecommerce"
	userID       = "demo_user"
	maxToolCalls = 4
)

type ChatRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

var toolMapping = map[string]string{
	"list_categories":         "Tìm kiếm danh mục sản phẩm phù hợp...",
	"get_specs_by_category":   "Kiểm tra thông số kỹ thuật (Hỗ trợ GPS, Chạy bộ)...",
	"list_products":           "Tìm kiếm các sản phẩm trong khoảng giá yêu cầu...",
	"get_product_by_id":       "Kiểm tra chi tiết thông tin và tồn kho sản phẩm...",
	"hybrid_search_documents": "Tìm kiếm tài liệu chính sách mua sắm/bảo hành...",
	"list_brands":             "Kiểm tra danh sách các thương hiệu...",
}

type ChatResponse struct {
	SessionID string   `json:"session_id"`
	Message   string   `json:"message"`
	Steps     []string `json:"steps,omitempty"`

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

type turnCounterKeyType string

const turnCounterKey turnCounterKeyType = "turn_counter"

type turnCounter struct {
	count int
}

type limitingLLM struct {
	model.LLM
	maxCalls int
}

func (l *limitingLLM) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	if tc, ok := ctx.Value(turnCounterKey).(*turnCounter); ok {
		if req.Config != nil && len(req.Config.Tools) > 0 {
			tc.count++
			if tc.count > l.maxCalls {
				log.Printf("[WARN] maxToolCalls (%d) exceeded, stripping tools and sanitizing history to force final response", l.maxCalls)
				cfgCopy := *req.Config
				cfgCopy.Tools = nil
				cfgCopy.ToolConfig = nil

				if cfgCopy.SystemInstruction == nil {
					cfgCopy.SystemInstruction = &genai.Content{
						Role:  "system",
						Parts: []*genai.Part{{Text: "The tool execution limit has been reached. You MUST summarize the results found so far and provide your final response to the user now. Do not attempt to call any tools."}},
					}
				} else {
					sysInstCopy := *cfgCopy.SystemInstruction
					sysInstCopy.Parts = append([]*genai.Part{}, sysInstCopy.Parts...)
					sysInstCopy.Parts = append(sysInstCopy.Parts, &genai.Part{
						Text: "\nIMPORTANT: The tool execution limit has been reached. You MUST summarize the results found so far and provide your final response to the user now. Do not attempt to call any tools.",
					})
					cfgCopy.SystemInstruction = &sysInstCopy
				}

				req.Config = &cfgCopy

				// Rewrite conversation history to convert tool calls/responses into text representation
				req.Contents = sanitizeHistoryForSummary(req.Contents)
			}
		}
	}
	return l.LLM.GenerateContent(ctx, req, stream)
}

func sanitizeHistoryForSummary(contents []*genai.Content) []*genai.Content {
	sanitized := make([]*genai.Content, 0, len(contents))
	for _, c := range contents {
		if c == nil {
			continue
		}
		newContent := &genai.Content{
			Role: c.Role,
		}
		// Map tool/function roles to user role to prevent breaking OpenAI-compatible APIs validation
		if c.Role == "tool" || c.Role == "function" {
			newContent.Role = string(genai.RoleUser)
		} else if c.Role == "model" {
			newContent.Role = string(genai.RoleModel)
		}

		for _, part := range c.Parts {
			if part == nil {
				continue
			}
			if part.FunctionCall != nil {
				argsJSON, _ := json.Marshal(part.FunctionCall.Args)
				newContent.Parts = append(newContent.Parts, &genai.Part{
					Text: fmt.Sprintf("[Hệ thống: Gọi tool %s với tham số %s]", part.FunctionCall.Name, string(argsJSON)),
				})
			} else if part.FunctionResponse != nil {
				respJSON, _ := json.Marshal(part.FunctionResponse.Response)
				newContent.Parts = append(newContent.Parts, &genai.Part{
					Text: fmt.Sprintf("[Hệ thống: Kết quả tool %s: %s]", part.FunctionResponse.Name, string(respJSON)),
				})
			} else if part.Text != "" {
				newContent.Parts = append(newContent.Parts, &genai.Part{
					Text: part.Text,
				})
			} else {
				newContent.Parts = append(newContent.Parts, part)
			}
		}
		sanitized = append(sanitized, newContent)
	}
	return sanitized
}

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

func NewAgentServer(ctx context.Context,
	appCfg *config.AppConfig,
	telemetry *observability.Telemetry,
	langfusePlg *runner.PluginConfig,
	llm model.LLM) (*AgentServer, error) {

	// 2. Init Shared Resources
	// Shared MCP Transport
	transport := &mcp.SSEClientTransport{
		Endpoint: appCfg.McpServer,
		HTTPClient: &http.Client{
			Transport: &headerTransport{
				base:   http.DefaultTransport,
				header: "Authorization",
				value:  "",
			},
		},
	}

	// Initialize Single Agent
	agentCfg, ok := appCfg.Agents["ecommerce_agent"]
	if !ok {
		return nil, fmt.Errorf("ecommerce_agent config not found")
	}

	// Initialize MCP tools for this agent
	mcpToolset, err := mymcp.NewMCPTool(transport, agentCfg.AllowedTools, agentCfg.ApprovedTools)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP tools for ecommerce_agent: %w", err)
	}

	wrappedLLM := &limitingLLM{
		LLM:      llm,
		maxCalls: maxToolCalls,
	}

	// Create the single Agent instance
	targetAgent, err := agents.NewSingleAgent(ctx, &agentCfg, &appCfg.Prompts, wrappedLLM, mcpToolset)
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
		PluginConfig:   *langfusePlg,
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

	sessionID := r.SessionID
	var exists bool
	if sessionID != "" {
		_, err := s.SessionService.Get(c.Request.Context(), &session.GetRequest{
			UserID:    userID,
			SessionID: sessionID,
			AppName:   appName,
		})
		if err == nil {
			exists = true
		}
	}

	if !exists {
		if sessionID == "" {
			sessionID = uuid.NewString()
		}
		_, err := s.SessionService.Create(c.Request.Context(), &session.CreateRequest{
			UserID:    userID,
			SessionID: sessionID,
			AppName:   appName,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Failed to create session": err.Error()})
			log.Printf("Failed to create session: %v", err)
			return
		}
	}
	r.SessionID = sessionID

	userMsg := genai.NewContentFromText(r.Message, genai.RoleUser)

	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, turnCounterKey, &turnCounter{})
	if authToken := c.GetHeader("Authorization"); authToken != "" {
		ctx = context.WithValue(ctx, contextKeyAuthToken, authToken)
	}

	// ctxOtel, span := s.Telemetry.Tracer.Start(ctx, "agent.request")
	ctxOtel, span := otel.Tracer("ecommerce-agent").Start(ctx, "agent.request")
	defer span.End()

	// Turn 1: chạy bình thường, capture confirmation event
	var pendingConfirmations map[string]toolconfirmation.ToolConfirmation
	var confirmationCallID string

	finalResponse := ""
	var steps []string

	for event, err := range s.Runner.Run(ctxOtel, userID, r.SessionID, userMsg, agent.RunConfig{}) {
		if err != nil {
			log.Printf("Run error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"Failed to run agent ": err.Error()})
			return
		}

		if event.Content != nil {
			for _, part := range event.Content.Parts {
				if part.Text != "" && event.Author != userID && event.Author != "user" {
					finalResponse += part.Text
				}
				if part.FunctionCall != nil {
					toolName := part.FunctionCall.Name
					if toolName == "adk_request_confirmation" {
						confirmationCallID = part.FunctionCall.ID
					} else {
						friendlyName, ok := toolMapping[toolName]
						if !ok {
							friendlyName = fmt.Sprintf("Đang xử lý bước %s...", toolName)
						}
						steps = append(steps, friendlyName)
					}
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
					Steps:                steps,
				})
				return
			}
		}
	}
	c.JSON(http.StatusOK, ChatResponse{
		SessionID: r.SessionID,
		Message:   finalResponse,
		Steps:     steps,
	})
}

func (s *AgentServer) HandlerConfirm(c *gin.Context) {
	var r ConfirmRequest
	if err := c.ShouldBindBodyWithJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"Invalid request": err.Error()})
		return
	}

	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, turnCounterKey, &turnCounter{})
	if authToken := c.GetHeader("Authorization"); authToken != "" {
		ctx = context.WithValue(ctx, contextKeyAuthToken, authToken)
	}

	ctxOtel, span := s.Telemetry.Tracer.Start(ctx, "agent.confirm")
	// ctxOtel, span := otel.Tracer("ecommerce-agent").Start(ctx, "agent.confirm")
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
	var steps []string

	for event, err := range s.Runner.Run(ctxOtel, userID, r.SessionID, approvalMsg, agent.RunConfig{}) {
		if err != nil {
			log.Printf("Resume error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if event.Content != nil {
			for _, part := range event.Content.Parts {
				if part.Text != "" && event.Author != userID && event.Author != "user" {
					finalResponse += part.Text
				}
				if part.FunctionCall != nil {
					toolName := part.FunctionCall.Name
					if toolName != "adk_request_confirmation" {
						friendlyName, ok := toolMapping[toolName]
						if !ok {
							friendlyName = fmt.Sprintf("Đang xử lý bước %s...", toolName)
						}
						steps = append(steps, friendlyName)
					}
				}
			}
		}
	}

	// Trả kết quả cuối cùng
	c.JSON(http.StatusOK, ChatResponse{
		SessionID: r.SessionID,
		Message:   finalResponse,
		Steps:     steps,
	})
}

func (s *AgentServer) HandlerChatStream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	var r ChatRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		c.SSEvent("error", map[string]string{"error": err.Error()})
		c.Writer.Flush()
		return
	}

	slog.Info("Chat request: ",
		"session_id", r.SessionID,
		"user_id", userID,
		"app_name", appName,
		"message", r.Message,
	)

	sessionID := r.SessionID
	var exists bool
	if sessionID != "" {
		_, err := s.SessionService.Get(c.Request.Context(), &session.GetRequest{
			UserID:    userID,
			SessionID: sessionID,
			AppName:   appName,
		})
		if err == nil {
			exists = true
		}
	}

	if !exists {
		if sessionID == "" {
			sessionID = uuid.NewString()
		}
		_, err := s.SessionService.Create(c.Request.Context(), &session.CreateRequest{
			UserID:    userID,
			SessionID: sessionID,
			AppName:   appName,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Failed to create session": err.Error()})
			log.Printf("Failed to create session: %v", err)
			return
		}
	}
	r.SessionID = sessionID

	userMsg := genai.NewContentFromText(r.Message, genai.RoleUser)
	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, turnCounterKey, &turnCounter{})
	if authToken := c.GetHeader("Authorization"); authToken != "" {
		ctx = context.WithValue(ctx, contextKeyAuthToken, authToken)
	}

	// ctxOtel, span := s.Telemetry.Tracer.Start(ctx, "agent.request.stream")
	ctxOtel, span := otel.Tracer("ecommerce-agent").Start(ctx, "agent.request.stream")
	defer span.End()

	var pendingConfirmations map[string]toolconfirmation.ToolConfirmation
	var confirmationCallID string

	// Send initial session ID
	c.SSEvent("session", map[string]string{"session_id": r.SessionID})
	c.Writer.Flush()

	var partialResponse string
	for event, err := range s.Runner.Run(ctxOtel, userID, r.SessionID, userMsg, agent.RunConfig{StreamingMode: "sse"}) {
		if err != nil {
			log.Printf("Run error: %v", err)
			if ctxOtel.Err() != nil {
				log.Printf("Session %s aborted by client. Saving partial response: %q", r.SessionID, partialResponse)
				s.saveCancelledSession(r.SessionID, userMsg, partialResponse)
			}
			c.SSEvent("error", map[string]string{"error": err.Error()})
			c.Writer.Flush()
			return
		}

		if event.Content != nil {
			for _, part := range event.Content.Parts {
				if part.Text != "" && event.Author != userID && event.Author != "user" {
					if event.Partial {
						partialResponse += part.Text
						c.SSEvent("token", map[string]string{"text": part.Text})
						c.Writer.Flush()
					}
				}
				if part.FunctionCall != nil {
					toolName := part.FunctionCall.Name
					if toolName == "adk_request_confirmation" {
						confirmationCallID = part.FunctionCall.ID
					} else {
						friendlyName, ok := toolMapping[toolName]
						if !ok {
							friendlyName = fmt.Sprintf("Đang xử lý bước %s...", toolName)
						}
						c.SSEvent("step", map[string]string{
							"message": friendlyName,
							"tool":    toolName,
						})
						c.Writer.Flush()
					}
				}
			}
		}
		if event.Actions.RequestedToolConfirmations != nil {
			pendingConfirmations = event.Actions.RequestedToolConfirmations
		}
	}

	if len(pendingConfirmations) > 0 {
		for callID, conf := range pendingConfirmations {
			if confirmationCallID != "" && callID != "" {
				c.SSEvent("confirmation", map[string]any{
					"confirmation_id": confirmationCallID,
					"hint":            conf.Hint,
				})
				c.Writer.Flush()
				return
			}
		}
	}

	c.SSEvent("done", map[string]string{"status": "completed"})
	c.Writer.Flush()
}

func (s *AgentServer) HandlerConfirmStream(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	var r ConfirmRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		c.SSEvent("error", map[string]string{"error": err.Error()})
		c.Writer.Flush()
		return
	}

	ctx := c.Request.Context()
	ctx = context.WithValue(ctx, turnCounterKey, &turnCounter{})
	if authToken := c.GetHeader("Authorization"); authToken != "" {
		ctx = context.WithValue(ctx, contextKeyAuthToken, authToken)
	}

	// ctxOtel, span := s.Telemetry.Tracer.Start(ctx, "agent.confirm.stream")
	ctxOtel, span := otel.Tracer("ecommerce-agent").Start(ctx, "agent.confirm.stream")
	defer span.End()

	confirmationCallID := r.ConfirmationID
	var parts []*genai.Part

	if confirmationCallID != "" {
		parts = append(parts, &genai.Part{
			FunctionResponse: &genai.FunctionResponse{
				ID:   confirmationCallID,
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

	var pendingConfirmations map[string]toolconfirmation.ToolConfirmation
	var nextConfirmationCallID string

	var partialResponse string
	for event, err := range s.Runner.Run(ctxOtel, userID, r.SessionID, approvalMsg, agent.RunConfig{StreamingMode: "sse"}) {
		if err != nil {
			log.Printf("Resume error: %v", err)
			if ctxOtel.Err() != nil {
				log.Printf("Session %s confirm stream aborted by client. Saving partial response: %q", r.SessionID, partialResponse)
				s.saveCancelledSession(r.SessionID, approvalMsg, partialResponse)
			}
			c.SSEvent("error", map[string]string{"error": err.Error()})
			c.Writer.Flush()
			return
		}

		if event.Content != nil {
			for _, part := range event.Content.Parts {
				if part.Text != "" && event.Author != userID && event.Author != "user" {
					if event.Partial {
						partialResponse += part.Text
						c.SSEvent("token", map[string]string{"text": part.Text})
						c.Writer.Flush()
					}
				}
				if part.FunctionCall != nil {
					toolName := part.FunctionCall.Name
					if toolName == "adk_request_confirmation" {
						nextConfirmationCallID = part.FunctionCall.ID
					} else {
						friendlyName, ok := toolMapping[toolName]
						if !ok {
							friendlyName = fmt.Sprintf("Đang xử lý bước %s...", toolName)
						}
						c.SSEvent("step", map[string]string{
							"message": friendlyName,
							"tool":    toolName,
						})
						c.Writer.Flush()
					}
				}
			}
		}

		if event.Actions.RequestedToolConfirmations != nil {
			pendingConfirmations = event.Actions.RequestedToolConfirmations
		}
	}

	if len(pendingConfirmations) > 0 {
		for callID, conf := range pendingConfirmations {
			if nextConfirmationCallID != "" && callID != "" {
				c.SSEvent("confirmation", map[string]any{
					"confirmation_id": nextConfirmationCallID,
					"hint":            conf.Hint,
				})
				c.Writer.Flush()
				return
			}
		}
	}

	c.SSEvent("done", map[string]string{"status": "completed"})
	c.Writer.Flush()
}

func (s *AgentServer) saveCancelledSession(sessionID string, userMsg *genai.Content, partialResponse string) {
	ctx := context.Background()
	sess, err := s.SessionService.Get(ctx, &session.GetRequest{
		UserID:    userID,
		SessionID: sessionID,
		AppName:   appName,
	})
	if err != nil {
		log.Printf("[WARN] Failed to fetch session %s to save partial response: %v", sessionID, err)
		return
	}

	// Append user message event
	userEvent := session.NewEvent("")
	userEvent.Author = userID
	userEvent.LLMResponse.Content = userMsg
	if err := s.SessionService.AppendEvent(ctx, sess.Session, userEvent); err != nil {
		log.Printf("[WARN] Failed to append user event on cancellation for session %s: %v", sessionID, err)
		return
	}

	// Append partial model response event if we have text
	if partialResponse != "" {
		modelEvent := session.NewEvent("")
		modelEvent.Author = "model"
		modelEvent.LLMResponse.Content = genai.NewContentFromText(partialResponse, genai.RoleModel)
		if err := s.SessionService.AppendEvent(ctx, sess.Session, modelEvent); err != nil {
			log.Printf("[WARN] Failed to append model partial response on cancellation for session %s: %v", sessionID, err)
		}
	}
	log.Printf("[INFO] Saved cancelled session %s successfully", sessionID)
}

func (s *AgentServer) HandlerDebugSession(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(400, gin.H{"error": "session_id query param is required"})
		return
	}

	sess, err := s.SessionService.Get(c.Request.Context(), &session.GetRequest{
		UserID:    userID,
		SessionID: sessionID,
		AppName:   appName,
	})
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	// 1. Trích xuất toàn bộ Events từ iterator của ADK
	var eventsList []*session.Event
	if sess.Session.Events() != nil {
		for ev := range sess.Session.Events().All() {
			eventsList = append(eventsList, ev)
		}
	}

	// 2. Trích xuất State hiện tại
	stateMap := make(map[string]any)
	if sess.Session.State() != nil {
		for k, v := range sess.Session.State().All() {
			stateMap[k] = v
		}
	}

	// 3. Trả về cấu trúc JSON đầy đủ của Session đang nằm trong Memory
	c.JSON(200, gin.H{
		"session_id": sess.Session.ID(),
		"app_name":   sess.Session.AppName(),
		"user_id":    sess.Session.UserID(),
		"updated_at": sess.Session.LastUpdateTime(),
		"state":      stateMap,
		"events":     eventsList,
	})
}
