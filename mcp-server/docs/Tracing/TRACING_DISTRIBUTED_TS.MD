## Tổng kết: Distributed Tracing với MCP Go SDK + Gin

### Vấn đề ban đầu
Gin tracing middleware tạo span `http.request` nhưng MCP SDK tạo span `mcp.request.*` riêng biệt, không có parent-child relationship. Trace bị gãy thành các trace độc lập trong Jaeger.

---

### Root Cause: Context propagation bị đứt

**Với SSE transport (`NewSSEHandler`):**

SDK có kiến trúc 2 connection:
- `GET /mcp/sse` → tạo session với context **tại thời điểm đó**
- `POST /mcp/message` → SDK lookup session cũ, dùng context của session, **bỏ qua** context của POST request

Dù Gin middleware đã inject span vào `c.Request.WithContext(ctx)` đúng cách, MCP SDK không forward context của từng POST request vào session pipeline. Span bị đứt là by design của SSE transport.

**Với `gin.WrapH()`:**

Gin lưu context vào `*gin.Context`, nhưng `gin.WrapH` forward `http.Request` gốc sang handler — không phải request đã được enrich. Fix là `c.Request = c.Request.WithContext(ctx)` trong middleware, nhưng với SSE vẫn không đủ vì vấn đề session context ở trên.

---

### Giải pháp: Chuyển sang StreamableHTTPHandler

`NewStreamableHTTPHandler` là transport mới hơn, stateless per-request — mỗi POST request xử lý độc lập với context của chính nó. Gin middleware inject span → `c.Request.WithContext(ctx)` → MCP SDK nhận đúng context → span có parent-child đúng.

```go
// mcp_handler.go
return mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
    return s
}, nil)
```

```go
// observation.go - key line
c.Request = c.Request.WithContext(ctx)  // bắt buộc phải có
```

---

### Limitation còn lại

Dù dùng `StreamableHTTPHandler`, SDK vẫn maintain session state. Các span MCP của cùng 1 session bị gom vào 1 trace dài (8m 26s trong ví dụ), trong khi mỗi POST request tạo thêm 1 `http.request` span riêng. Không có cách clean để link 2 trace này thành parent-child vì SDK không expose hook để override context per-request sau khi session tạo.

Baggage cũng không giải quyết được vì cùng lý do — context bị đứt tại tầng session.

---

### Những gì vẫn hoạt động tốt

Tracing **bên trong tool** vẫn có đầy đủ giá trị. Inject tracer vào tool lúc khởi tạo, tạo span con từ `ctx` nhận được trong handler:

```go
func NewSearchTool(db database.Store, telemetry *observability.Telemetry) *SearchTool {
    return &SearchTool{db: db, tracer: telemetry.Tracer}
}

func (t *SearchTool) Handler(ctx context.Context, ...) {
    ctx, span := t.tracer.Start(ctx, "tool.search")
    defer span.End()
    // ctx này truyền xuống DB, Redis → tạo span con đúng
    results, err := t.db.Search(ctx, args.Query)
}
```

Chain trong Jaeger: `mcp.request.tools/call` → `tool.search` → `db.query` — đủ để debug performance và error bên trong tool.

---

### Correlate với upstream service

Vì không thể link HTTP trace với MCP session trace qua span hierarchy, giải pháp thực tế là dùng `session.id` làm correlation key:

- Trong Jaeger: search theo attribute `session.id`
- Upstream service log `session.id` ra structured log
- Dùng log aggregation (Loki, CloudWatch) để join nếu cần

---

### Bài học chính

| | SSE Transport | Streamable HTTP |
|---|---|---|
| Context propagation | Bị đứt tại session layer | Hoạt động đúng |
| Gin middleware inject | Không đủ | Đủ với `WithContext` |
| Per-request tracing | Không khả thi | Có parent-child |
| Status | Legacy | Recommended |

**Khi dùng SDK third-party với HTTP framework:** luôn kiểm tra SDK lấy context từ đâu — từ `r.Context()` của từng request hay từ một session/connection context được tạo một lần. Đó là điểm quyết định toàn bộ tracing strategy.