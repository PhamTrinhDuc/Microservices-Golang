# Tài liệu Tích hợp Xác thực Keycloak OIDC

Tài liệu này mô tả chi tiết luồng hoạt động xác thực (Authentication) và phân quyền (Authorization) sử dụng Keycloak OIDC cho hệ thống của chúng ta.

---

## 1. Luồng đăng nhập thông thường (Direct Access Grant / Password Flow)
Luồng này thường được sử dụng cho việc kiểm thử tự động, công cụ CLI hoặc đăng nhập trực tiếp qua API.

```mermaid
sequenceDiagram
    autonumber
    actor User as Người dùng
    participant FE as Frontend / Client App
    participant KC as Keycloak Auth Service
    participant DB as Keycloak Database

    User->>FE: Nhập Username/Password
    FE->>KC: POST /realms/multi-agent/protocol/openid-connect/token (username, password, client_id)
    Note over KC: Keycloak xác thực thông tin đăng nhập
    KC->>DB: Truy vấn thông tin người dùng
    DB-->>KC: Trả về thông tin (Password Hash, Roles)
    Note over KC: So sánh Hash mật khẩu & Tạo Access/Refresh Token (JWT)
    KC-->>FE: Trả về Access Token + Refresh Token (JSON)
    FE->>User: Đăng nhập thành công, hiển thị giao diện chính
```

---

## 2. Luồng đăng nhập bằng Google (OIDC Federation)
Luồng này sử dụng phương thức Authorization Code Flow phối hợp giữa Keycloak (như một Broker) và Google Identity Provider.

```mermaid
sequenceDiagram
    autonumber
    actor User as Người dùng
    participant FE as Frontend / Client App
    participant KC as Keycloak Auth Service
    participant GG as Google Identity Provider
    participant DB as Keycloak Database

    User->>FE: Click "Sign in with Google"
    FE->>KC: Chuyển hướng sang trang đăng nhập Keycloak (OAuth2 Auth Code flow)
    KC-->>User: Hiển thị trang login của Keycloak (có nút Login Google)
    User->>KC: Click "Login with Google"
    KC->>GG: Chuyển hướng sang Google Authentication
    User->>GG: Đăng nhập & Xác nhận cấp quyền tài khoản Google
    GG-->>KC: Trả về mã Authorization Code (Redirect về Keycloak)
    KC->>GG: Gửi Auth Code đổi lấy Google ID Token & Email
    GG-->>KC: Trả về thông tin tài khoản Google
    Note over KC: Kiểm tra tài khoản trong Keycloak DB
    alt Tài khoản chưa tồn tại
        KC->>DB: Đăng ký user mới và liên kết với Google ID
    else Tài khoản đã tồn tại
        KC->>DB: Cập nhật thông tin phiên đăng nhập
    end
    DB-->>KC: Xác nhận lưu trữ
    KC-->>FE: Chuyển hướng về Frontend kèm Authorization Code của Keycloak
    FE->>KC: Gửi Auth Code đổi lấy Access Token & Refresh Token của hệ thống
    KC-->>FE: Trả về Access Token (JWT)
```

---

## 3. Luồng gọi API có xác thực (API Gateway & Microservices Verification)
Luồng này mô tả cách các dịch vụ (`backend`, `multi-agent`, `mcp-server`) xác thực token được gửi từ Frontend, đồng thời tự động đồng bộ tài khoản người dùng vào DB local của Backend.

```mermaid
sequenceDiagram
    autonumber
    participant FE as Frontend Client
    participant MA as Multi-Agent
    participant MCP as MCP Server
    participant BE as Backend Service
    participant KC as Keycloak (JWKS)
    participant PG as Local Postgres DB

    FE->>MA: POST /api/chat (Header: Authorization: Bearer <token>)
    Note over MA: Đọc token và đính kèm vào Request Context

    MA->>MCP: Gọi Tool API qua SSE (Header: Authorization: Bearer <token>)
    Note over MCP: Verify Token
    opt Chưa cache Keycloak Public Keys
        MCP->>KC: GET /realms/multi-agent/protocol/openid-connect/certs (JWKS)
        KC-->>MCP: Trả về danh sách Public Keys để verify signature
    end
    Note over MCP: Xác minh chữ ký JWT bằng Public Key

    MCP->>BE: Gọi DB/API nội bộ (Header: Authorization: Bearer <token>)
    
    Note over BE: Verify Token (Auth Middleware)
    opt Chưa cache Keycloak Public Keys
        BE->>KC: GET /realms/multi-agent/protocol/openid-connect/certs
        KC-->>BE: Trả về Public Keys
    end
    Note over BE: Xác minh chữ ký JWT thành công

    Note over BE: Đồng bộ hóa User (getOrCreateUserID)
    BE->>PG: SELECT id FROM users WHERE keycloak_id = sub (Keycloak UUID)
    alt User đã tồn tại trong Postgres
        PG-->>BE: Trả về user_id (Integer)
    else User chưa tồn tại trong Postgres
        BE->>PG: SELECT id FROM users WHERE email = token.email
        alt Có User trùng email
            BE->>PG: UPDATE users SET keycloak_id = sub WHERE id = ...
        else Không trùng email
            BE->>PG: INSERT INTO users (keycloak_id, email, full_name)
        end
        PG-->>BE: Trả về user_id (Integer) mới hoặc đã map
    end
    
    Note over BE: Đưa user_id (Integer) và role vào Gin Context
    BE->>PG: Thực hiện nghiệp vụ lưu trữ (Carts, Orders...) bằng user_id (Integer)
    PG-->>BE: Kết quả truy vấn
    BE-->>MCP: Trả về dữ liệu Catalog/Order
    MCP-->>MA: Trả về kết quả Tool
    MA-->>FE: Phản hồi nội dung chat của Agent cho người dùng
```
