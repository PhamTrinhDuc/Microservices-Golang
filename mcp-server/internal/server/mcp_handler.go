package server

// https://zread.ai/modelcontextprotocol/go-sdk/11-middleware-and-request-handling
import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mcp-server/internal/database"
	"mcp-server/internal/observability"
	"mcp-server/internal/tools"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// tracingMiddleware dùng đúng type mcp.Middleware (không generic) của v1.x
func tracingMiddleware(telemetry *observability.Telemetry) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			if telemetry == nil || telemetry.Tracer == nil {
				return next(ctx, method, req)
			}

			startTime := time.Now()

			// 1. Khởi tạo Span
			ctx, span := telemetry.Tracer.Start(
				ctx, "mcp.request."+method,
				trace.WithAttributes(
					attribute.String("rpc.method", method),
					attribute.String("session.id", req.GetSession().ID()),
				),
			)
			defer span.End()

			// 2. Record Active Requests
			if telemetry.Metrics != nil {
				telemetry.Metrics.ActiveRequests.Add(ctx, 1)
				defer telemetry.Metrics.ActiveRequests.Add(ctx, -1)
			}

			// 3. Thực thi handler tiếp theo
			result, err := next(ctx, method, req)

			// 4. Xử lý Metrics và Span Status
			duration := time.Since(startTime)
			status := "success"

			// Marshal request sang JSON để log và trace
			reqBytes, _ := json.Marshal(req)
			reqStr := string(reqBytes)

			if err != nil {
				status = "error"
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
				log.Printf("[MCP ERROR] Method: %s | Req: %s | Error: %v\n", method, reqStr, err)
			} else {
				// Kiểm tra application-level error trong CallToolResult
				if r, ok := result.(*mcp.CallToolResult); ok && r != nil {
					if r.IsError {
						status = "error"
						// Lấy chi tiết nội dung lỗi từ content của tool
						errText := "unknown tool error"
						if len(r.Content) > 0 {
							contentBytes, errMarshal := json.Marshal(r.Content)
							if errMarshal == nil {
								errText = string(contentBytes)
							}
						}
						span.SetStatus(codes.Error, fmt.Sprintf("tool error: %s", errText))
						span.SetAttributes(
							attribute.Bool("tool.error", true),
							attribute.String("tool.error_message", errText),
						)
						log.Printf("[MCP TOOL ERROR] Method: %s | Req: %s | Result Error: %s\n", method, reqStr, errText)
					} else {
						span.SetStatus(codes.Ok, "success")
						log.Printf("[MCP SUCCESS] Method: %s | Req: %s | Duration: %v\n", method, reqStr, duration)
					}
				} else {
					span.SetStatus(codes.Ok, "success")
				}
			}

			if telemetry.Metrics != nil {
				telemetry.Metrics.RecordRequest(ctx, method, status, float64(duration.Milliseconds()))
			}
			return result, err
		}
	}
}

func NewSSEHandler(db database.Store, telemetry *observability.Telemetry, backendURL string) http.Handler {
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "a2a-mcp-server",
			Version: "1.0.0",
		},
		nil,
	)

	// Add receiving middleware for incoming responses
	s.AddReceivingMiddleware(tracingMiddleware(telemetry))

	hybridTool := tools.NewHybridSearchTool(db)
	categoryTool := tools.NewCategoryTool(backendURL)
	brandTool := tools.NewBrandTool(backendURL)
	productTool := tools.NewProductTool(backendURL)

	// Search document
	mcp.AddTool(s, hybridTool.Definition(), hybridTool.Handler)

	// API Category
	mcp.AddTool(s, categoryTool.ListCategoryDefinition(), categoryTool.ListCategoryHandler)

	// API Brand
	mcp.AddTool(s, brandTool.ListBrandDefinition(), brandTool.ListBrandHandler)

	// API Product
	mcp.AddTool(s, productTool.ListProductDefinition(), productTool.ListProductHandler)
	mcp.AddTool(s, productTool.GetProductDefinition(), productTool.GetProductHandler)
	mcp.AddTool(s, productTool.GetSpectByCategoryDefinition(), productTool.GetSpectByCategoryHandler)

	sseHandler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
		return s
	}, nil)

	// Wrap để propagate span từ Gin vào MCP
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// r.Context() tại đây đã có span của Gin (vì gin đã set WithContext)
		// nhưng cần inject vào MCP session context thông qua propagator
		ctx := r.Context()
		propagator := otel.GetTextMapPropagator()

		// Re-inject span vào header để MCP SDK extract được
		propagator.Inject(ctx, propagation.HeaderCarrier(r.Header))

		sseHandler.ServeHTTP(w, r)
	})
}

// func NewStreamableHTTPHandler(db database.Store, telemetry *observability.Telemetry) http.Handler {
// 	s := mcp.NewServer(
// 		&mcp.Implementation{
// 			Name:    "a2a-mcp-server",
// 			Version: "1.0.0",
// 		},
// 		nil,
// 	)

// 	s.AddReceivingMiddleware(tracingMiddleware(telemetry))

// 	searchTool := tools.NewSearchTool(db)
// 	hybridTool := tools.NewHybridSearchTool(db)

// 	mcp.AddTool(s, searchTool.Definition(), searchTool.Handler)
// 	mcp.AddTool(s, hybridTool.Definition(), hybridTool.Handler)

// 	// StreamableHTTPHandler dùng context của từng HTTP request
// 	// nên span từ Gin sẽ được propagate đúng
// 	return mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
// 		return s
// 	}, nil)
// }
