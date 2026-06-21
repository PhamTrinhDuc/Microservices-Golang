# Distributed Tracing Flow: Agent ➔ MCP Server ➔ Backend

Tài liệu này giải thích cách hệ thống truyền dẫn **Trace Context** xuyên suốt qua 3 lớp dịch vụ để đảm bảo một yêu cầu từ người dùng có thể được theo vết (trace) từ đầu đến cuối.

## 1. Tổng quan kiến trúc
Luồng di chuyển của Trace ID:
`Agent (salon-agent)` ➔ `MCP Server (a2a-mcp-server)` ➔ `Backend API (salon-backend)`

---

## 2. Chi tiết các bước thực hiện

### Bước 1: Khởi tạo (Root Span tại Agent)
Tại file `test.go`, chúng ta khởi tạo Telemetry và bắt đầu một Span cha.

*   **Hành động**: `telemetry.Tracer.Start(ctx, "agent-server")`
*   **Kết quả**: Tạo ra một `TraceID` duy nhất và lưu vào trong biến `ctxOtel`.
*   **Mục đích**: Đây là "gốc" của toàn bộ hành trình.

### Bước 2: Ký gửi (Injection tại Agent)
Trước khi request HTTP bay đi từ Agent sang MCP Server, nó phải đi qua `RoundTrip` trong `headerTransport`.

*   **Hành động**: `propagator.Inject(req.Context(), propagation.HeaderCarrier(newReq.Header))`
*   **Kết quả**: Header HTTP của request sẽ có thêm dòng:
    `traceparent: 00-{TraceID}-{SpanID}-01`
*   **Mục đích**: Mang ID từ trong code dán lên "phong bì" HTTP để gửi qua mạng.

### Bước 3: Tiếp nhận (Extraction tại MCP Server)
Khi MCP Server nhận được request, Middleware của Gin sẽ thực hiện giải mã.

*   **Hành động**: `otel.GetTextMapPropagator().Extract(ctx, ...)` (thực hiện ngầm trong middleware).
*   **Kết quả**: MCP Server lấy lại được `TraceID` gốc từ Header và đưa vào `context` của Go.
*   **Mục đích**: Nối lại sợi dây tracing đã bị ngắt quãng bởi môi trường mạng.

### Bước 4: Nối dài (Child Span tại MCP Server)
Trong `mcp_handler.go`, chúng ta tạo thêm một Span con cho từng tool cụ thể.

*   **Hành động**: `ctx, span := telemetry.Tracer.Start(ctx, "mcp.request."+method, ...)`
*   **Kết quả**: Một Span con được tạo ra, nằm dưới Span cha từ Agent.
*   **Mục đích**: Ghi lại chi tiết thời gian xử lý tại MCP Server.

### Bước 5: Tiếp tục ký gửi (Injection sang Backend)
Trong các Tool Handler (như `ListBranchHandler`), MCP Server lại gọi sang Backend.

*   **Hành động**: `propagator.Inject(ctx, propagation.HeaderCarrier(httpReq.Header))`
*   **Kết quả**: Dán tiếp `TraceID` gốc vào request gửi sang Backend API.
*   **Mục đích**: Đảm bảo Backend cũng nhận được cùng một ID đó.

---

## 3. Các quy tắc "vàng" để không mất Trace
1.  **Luôn truyền `ctx`**: Không bao giờ được dùng `context.Background()` ở giữa luồng. Nếu bác dùng `Background()`, sợi dây trace sẽ bị đứt ngay lập tức vì nó không chứa bất kỳ ID nào.
2.  **Sử dụng `NewRequestWithContext`**: Khi gọi HTTP, luôn phải dùng hàm này để gắn `ctx` vào request.
3.  **Inject trước khi gửi**: Luôn gọi `propagator.Inject` ngay trước khi thực hiện `client.Do()`.

---

## 4. Cách kiểm tra
Bác có thể mở giao diện Tracing (Jaeger/Zipkin) và tìm kiếm Service `salon-agent`. Bác sẽ thấy một biểu đồ hình thác nước (Waterfall) hiển thị chi tiết:
1.  Thời gian Agent xử lý.
2.  Thời gian MCP Server chạy Tool.
3.  Thời gian Backend API truy vấn Database.

---
> [!TIP]
> Việc bác sử dụng chung bộ `internal/observability` giữa các service là một lựa chọn rất tốt, giúp đảm bảo tính đồng nhất về mặt kỹ thuật (như cách dùng chung Propagator TraceContext và Baggage).
