# Luồng Xác Thực (Authentication Flow) & Truyền Token trong Multi-Agent System

Tài liệu này mô tả chi tiết thiết kế bảo mật cho hệ thống Multi-Agent, cách JWT Token của người dùng (User) được truyền tự động và an toàn qua nhiều lớp dịch vụ: từ **Client (Frontend)** $\rightarrow$ **Agent Server** $\rightarrow$ **MCP Server** $\rightarrow$ **Backend**.

---

## 1. Biểu Đồ Trực Quan (Mermaid Sequence Diagram)

```mermaid
sequenceDiagram
    autonumber
    actor Client
    participant Agent_API as frontend/agentAPI
    participant Agent_Server as Agent Server (Gin)
    participant ADK_Runner as ADK Runner (Agent)
    participant MCP_SDK as MCP Client SDK
    participant RT as headerTransport (RoundTripper)
    participant MCP_Server as MCP Server (Gin)
    participant Backend as Backend System

    %% Kịch bản A: Khởi động hệ thống
    Note over Agent_Server, MCP_Server: [KỊCH BẢN A: KHỞI ĐỘNG HỆ THỐNG - STARTUP]
    Agent_Server->>MCP_Server: 1. Khởi tạo SSE & Gọi ListTools (KHÔNG có Token)
    Note over MCP_Server: Dùng authMid.OptionalHandler()<br/>Cho phép request đi qua
    MCP_Server-->>Agent_Server: Trả về danh sách cấu trúc Tools (Public Metadata)
    Note over Agent_Server: Đăng ký cấu trúc tool với LLM thành công

    %% Kịch bản B: Người dùng thực thi Chat
    Note over Client, Backend: [KỊCH BẢN B: RUNTIME - USER CHAT & GỌI TOOL]
    Client->>Agent_API: Nhập tin nhắn cần Chat
    Agent_API->>Agent_Server: 2. POST /api/chat<br/>Header: Authorization: Bearer <User Token>
    
    rect rgb(220, 240, 255)
        Note over Agent_Server: Trích xuất token từ Header<br/>ctx = context.WithValue(ctx, "auth_token", token)
    end
    
    Agent_Server->>ADK_Runner: s.Runner.Run(ctx, ...)
    Note over ADK_Runner: Agent phân tích & quyết định chạy Tool
    
    ADK_Runner->>MCP_SDK: Gọi Tool (Truyền ctx mang token)
    MCP_SDK->>RT: Thực hiện HTTP Request (sseClient.Do)
    
    rect rgb(230, 255, 230)
        Note over RT: req.Context().Value("auth_token") -> Token<br/>newReq.Header.Set("Authorization", Token)
    end
    
    RT->>MCP_Server: 3. POST /mcp/session/tools/call<br/>Header: Authorization: Bearer <User Token>
    
    rect rgb(255, 240, 220)
        Note over MCP_Server: authMid.OptionalHandler() phát hiện có Token<br/>Xác thực token JWT & lưu user_id vào context
        Note over MCP_Server: MCP Middleware chặn check: method == "tools/call"<br/>Bắt buộc phải có user_id (Đã được Authenticated)
    end
    
    MCP_Server->>Backend: 4. Gọi API Backend (Forward Token nếu cần)
    Backend-->>MCP_Server: Trả về kết quả nghiệp vụ (Data)
    MCP_Server-->>RT: Trả kết quả tool
    RT-->>MCP_SDK: Trả kết quả HTTP Response
    MCP_SDK-->>ADK_Runner: Trả kết quả về Agent
    Note over ADK_Runner: LLM tổng hợp dữ liệu & sinh câu trả lời
    ADK_Runner-->>Agent_Server: Hoàn tất hội thoại
    Agent_Server-->>Client: Trả về ChatResponse cho User
```

---

## 2. Chi Tiết Thiết Kế Từng Lớp

### Lớp A: Client (Frontend)
* **File liên quan:** [agentAPI.ts](file:///d:/Project/Go-Project/frontend/src/services/agentAPI.ts)
* **Nhiệm vụ:**
  * Lấy JWT token hiện tại của user từ `tokenManager.getToken()`.
  * Chèn Token vào header `Authorization: Bearer <token>` cho mọi cuộc gọi chat (`/api/chat` và `/api/chat/confirm`).

### Lớp B: Agent Server (Bố trí Context Propagation)
* **File liên quan:** [agent_handler.go](file:///d:/Project/Go-Project/single-agent/server/agent_handler.go)
* **Nhiệm vụ:**
  1. **Trích xuất Token:** Đọc `Authorization` từ HTTP Header gửi tới, lưu vào `context.Context` thông qua một khóa định danh riêng (`contextKeyAuthToken`).
  2. **Truyền dọc Context:** Chuyển tiếp Context này vào `Runner.Run(ctx, ...)` của ADK. Khi ADK gọi MCP SDK, context này sẽ tự động được mang theo.
  3. **Tiêm Token động tại RoundTripper:** Đối tượng `headerTransport` triển khai `http.RoundTripper` chặn mọi request đi ra từ MCP SDK sang MCP Server. Nếu phát hiện `auth_token` trong Context của request, nó sẽ đè lại header `Authorization` bằng Token thật của User hiện tại.

### Lớp C: MCP Server (Bảo vệ theo Phân Quyền Endpoint)
* **File liên quan:** [main.go](file:///d:/Project/Go-Project/mcp-server/cmd/main.go) & [mcp_handler.go](file:///d:/Project/Go-Project/mcp-server/internal/server/mcp_handler.go)
* **Nhiệm vụ:**
  1. **Xác thực tùy chọn (Optional Authentication):** Route `/mcp` sử dụng `OptionalHandler()`. 
     * Nếu không truyền token (lúc khởi động hệ thống từ Agent Server), request vẫn đi qua để lấy danh sách Tool mô tả (Public Metadata).
     * Nếu có truyền token, tiến hành giải mã và xác thực chữ ký JWT. Nếu đúng, lưu `user_id` vào Context.
  2. **Kiểm tra nghiệp vụ (Action Authorization):** Một Middleware MCP kiểm soát tất cả các lệnh gọi.
     * Khi nhận được phương thức `"tools/call"` (thực thi hành động thật), middleware yêu cầu bắt buộc phải tồn tại `user_id` trong context.
     * Nếu không có (do token không hợp lệ hoặc thiếu), từ chối ngay lập tức bằng lỗi `401/Unauthorized`.

---

## 3. Các Ưu Điểm Của Thiết Kế Này (Trade-offs)
* **Không dùng System/Admin Token tĩnh:** Tránh hoàn toàn việc lưu cấu hình Token hệ thống dài hạn ở file cấu hình `.env` của backend. Mọi hành động thao tác dữ liệu đều chạy dưới danh nghĩa (on behalf of) của chính người dùng hiện tại.
* **Tối ưu hóa Startup:** Việc tách biệt quyền của `ListTools` giúp Agent Server khởi chạy mượt mà ngay lập tức mà không cần phụ thuộc vào bất kỳ phiên đăng nhập nào của Client.
* **An toàn tuyệt đối:** Chỉ cho phép xem cấu trúc của tool (public schema) mà không cho phép bất kỳ ai chạy tool lấy dữ liệu nếu không cung cấp một Token hợp lệ đã được xác thực qua Keycloak.
