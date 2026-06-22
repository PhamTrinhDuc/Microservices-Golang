## JWT Validation 
# auth — JWT Authentication Package

Package `auth` cung cấp middleware xác thực JWT dùng RSA public/private key, thiết kế cho hệ thống multi-tenant.

---

## Tổng quan

Package này chia làm 2 nhóm chức năng chính:

- **Xác thực token** — parse, verify chữ ký RSA, kiểm tra issuer/audience/expiry
- **Cung cấp thông tin auth** — lưu và trích xuất claims từ `context` để dùng trong handler

---

## Cài đặt

```bash
go get github.com/golang-jwt/jwt/v5
```

---

## Cấu trúc

```
auth/
└── auth.go       # JWTValidator, Claims, context helpers, token generator
```

---

## Sử dụng

### 1. Khởi tạo validator

```go
validator, err := auth.NewJWTValidator(auth.Config{
    PublicKeyPEM: "-----BEGIN PUBLIC KEY-----\n...",
    Issuer:       "mcp-server-demo",
    Audience:     "mcp-server",
})
```

### 2. Dùng trong middleware

```go
func AuthMiddleware(v *auth.JWTValidator) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")

            claims, err := v.ValidateToken(token)
            if err != nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            ctx := auth.WithAuth(r.Context(), claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### 3. Trích xuất thông tin trong handler

```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
    tenantID, err := auth.ExtractTenantID(r.Context())
    userID, err   := auth.ExtractUserID(r.Context())

    if !auth.HasScope(r.Context(), "admin") {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // business logic...
}
```

---

## API

### `NewJWTValidator(cfg Config) (*JWTValidator, error)`

- Khởi tạo validator từ RSA public key PEM. Server chỉ cần public key để verify — private key nằm riêng ở Auth Service.
- Trả về con trỏ `JWTValidator` dùng để validate token sau này. Không tạo bản sao của `JWTValidator` mỗi lần sử dụng sau này.

### `ValidateToken(tokenString string) (*Claims, error)`

Xác thực token qua 5 bước theo thứ tự:

1. Strip prefix `Bearer `
2. Parse và verify chữ ký RSA
3. Kiểm tra `issuer`
4. Kiểm tra `audience`
5. Kiểm tra `tenant_id` không rỗng

### `WithAuth(ctx, claims) context.Context`

Gắn `tenant_id`, `user_id`, `scopes` vào context. Gọi ở middleware sau khi validate xong.

### `ExtractTenantID(ctx) (string, error)`
### `ExtractUserID(ctx) (string, error)`
### `ExtractScopes(ctx) ([]string, error)`

Trích xuất thông tin auth từ context. Trả về lỗi nếu context không có giá trị tương ứng.

### `HasScope(ctx, scope) bool`

Kiểm tra nhanh xem user có scope cụ thể không. Trả về `false` nếu không tìm thấy (không panic).

---

## Cấu trúc Claims

| Field | JSON key | Bắt buộc | Mô tả |
|---|---|---|---|
| `TenantID` | `tenant_id` | Có | ID tenant, dùng để phân tách dữ liệu |
| `UserID` | `user_id` | Có | ID người dùng |
| `Email` | `email` | Không | Email, bỏ qua nếu rỗng |
| `Scopes` | `scopes` | Không | Danh sách quyền hạn |

Ngoài ra embed `jwt.RegisteredClaims` gồm các field chuẩn: `exp`, `iss`, `aud`, `iat`, `nbf`.

---

## Testing

Package cung cấp 2 hàm sinh token dùng cho môi trường dev/test. **Không dùng ở production.**

```go
// Token mặc định, hết hạn sau 24h
token, err := auth.GenerateDemoToken("tenant-123", "user-456", []string{"read"}, privateKey)

// Tùy chỉnh thời gian hết hạn — dùng để test edge case
expiredToken, err := auth.GenerateDemoTokenWithExpiry(
    "tenant-123", "user-456", []string{},
    privateKey,
    -1 * time.Hour,   // token đã hết hạn
)
```

> **Lý do không dùng ở production:** Hàm này cần private key. Ở production, private key phải nằm riêng ở Auth Service — server ứng dụng chỉ được giữ public key để verify.

---

## Luồng xác thực

```
HTTP Request (Authorization: Bearer <token>)
    │
    ▼
Middleware: ValidateToken()
    ├── Thất bại → 401 Unauthorized
    └── Thành công → WithAuth(ctx, claims)
                         │
                         ▼
                    Business Handler
                         │
                         ├── ExtractTenantID(ctx)
                         ├── ExtractUserID(ctx)
                         └── HasScope(ctx, "admin")
```