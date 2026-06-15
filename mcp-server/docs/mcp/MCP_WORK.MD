Đây là luồng hoạt động đầy đủ của MCP SSE transport:

```
┌─────────────────────────────────────────────────────────────────┐
│                     LUỒNG HOẠT ĐỘNG MCP SSE                     │
└─────────────────────────────────────────────────────────────────┘

BƯỚC 1: Client mở SSE connection
──────────────────────────────────
Client                                    Server
  │                                          │
  │  GET /mcp/sse                            │
  │  Accept: text/event-stream               │
  │─────────────────────────────────────────►│
  │                                          │ tạo session
  │◄─────────────────────────────────────────│
  │  event: endpoint                         │
  │  data: /mcp/message?sessionId=abc123     │ ← server gửi về URL để POST
  │                                          │
  │  (connection giữ mở, chờ event tiếp...)  │


BƯỚC 2: Client gửi JSON-RPC request qua POST
─────────────────────────────────────────────
Client                                    Server
  │                                          │
  │  POST /mcp/message?sessionId=abc123      │
  │  {"jsonrpc":"2.0",                       │
  │   "method":"tools/list", "id":1}         │
  │─────────────────────────────────────────►│
  │                                          │ xử lý
  │  HTTP 202 Accepted                       │
  │◄─────────────────────────────────────────│


BƯỚC 3: Server trả kết quả qua SSE stream (bước 1 vẫn mở)
───────────────────────────────────────────────────────────
Client                                    Server
  │                                          │
  │◄─────────────────────────────────────────│
  │  event: message                          │
  │  data: {"jsonrpc":"2.0",                 │
  │          "result":{...tools...},         │
  │          "id":1}                         │
  │                                          │
```

Vì vậy **client cần làm 2 việc song song:**

```
Kết nối 1 (SSE - giữ mở):   GET /mcp/sse
                                  │
                                  └── nhận sessionId
                                  └── lắng nghe response

Kết nối 2 (HTTP thường):    POST /mcp/message?sessionId=xxx
                                  │
                                  └── gửi request
                                  └── nhận 202, kết quả về qua kết nối 1
```

Dùng MCP Inspector để test trực quan nhất:

```bash
npx @modelcontextprotocol/inspector

# Điền vào form:
# Transport: SSE
# URL: http://localhost:8080/mcp/sse
# Header: Authorization: Bearer <token>
```

Hoặc test thủ công bằng curl trong 2 terminal:

```bash
# Terminal 1 - mở SSE, chờ nhận sessionId và response
curl -N \
  -H "Authorization: Bearer <token>" \
  -H "Accept: text/event-stream" \
  http://localhost:8080/mcp/sse

# → server trả về:
# event: endpoint
# data: /mcp/message?sessionId=abc123


# Terminal 2 - gửi request với sessionId vừa nhận
curl -X POST \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}' \
  "http://localhost:8080/mcp/message?sessionId=abc123"

# → Terminal 1 sẽ nhận được:
# event: message
# data: {"jsonrpc":"2.0","result":{...},"id":1}
```

Đó là lý do tại sao `/mcp/health` bị treo — SSE handler **không bao giờ đóng connection**, nó thiết kế để giữ mở chờ event.