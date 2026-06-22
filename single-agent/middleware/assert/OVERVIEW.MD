## 1. Auth Middleware 
### Luồng hoàn chỉnh: HTTP Request → Database

Trước tiên hãy xem middleware này nằm ở đâu trong toàn bộ hệ thống:

```
HTTP Request
     │
     ▼
┌─────────────────────┐
│   AuthMiddleware    │  ← file này
│  (kiểm tra token)   │
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│    HTTP Handler     │
│  (xử lý business)   │
└─────────────────────┘
     │
     ▼
┌─────────────────────┐
│     PostgreSQL      │
│   (Row Level Sec)   │
└─────────────────────┘
```

---

### Luồng chi tiết theo từng hàm

#### 1. `Handler` — xác thực bắt buộc
```
Request đến
    │
    ├─ Không có Authorization header → 401 ngay, dừng
    │
    ├─ Có header → ValidateToken()
    │       ├─ Token sai/hết hạn → 401, dừng
    │       └─ Token hợp lệ → lấy được Claims {TenantID, UserID, Scopes}
    │
    └─ WithAuth(ctx, claims) → nhét Claims vào context
            │
            └─ next.ServeHTTP() → đi tiếp vào handler
```

#### 2. `OptionalHandler` — xác thực tùy chọn
```
Request đến
    │
    ├─ Không có header → đi tiếp KHÔNG có auth context (guest)
    │
    └─ Có header → ValidateToken()
            ├─ Token hợp lệ → WithAuth() → đi tiếp CÓ auth context
            └─ Token sai → 401, dừng  ← "nếu đã cố gửi thì phải đúng"
```

> Cái hay ở đây: **không gửi thì được, gửi sai thì không tha**. Dành cho những endpoint như `initialize` mà khách chưa có token vẫn cần gọi được.

---

### Kết nối với 2 file trước

Đây là luồng đầy đủ khi ghép cả 3 file lại:

```
① Client gửi request + Bearer token
        │
        ▼
② AuthMiddleware.Handler() — file middleware này
   └─ validator.ValidateToken(token)  → dùng JWTValidator (file auth)
           │
           └─ Lấy được claims.TenantID = "tenant_ABC"
                   │
                   └─ auth.WithAuth(ctx, claims)
                           │
                           ▼
③ Handler nhận request với ctx đã có TenantID
   └─ auth.ExtractTenantID(ctx) → "tenant_ABC"
           │
           └─ db.ExecTx(ctx, func(tx) {
                   │
                   ▼
④ setTenantContext(tx)  — file database
   └─ SET LOCAL app.current_tenant_id = 'tenant_ABC'
           │
           ▼
⑤ PostgreSQL tự động áp RLS
   └─ Mọi query trong tx chỉ thấy data của tenant_ABC
```

---

## Một điểm tinh tế đáng chú ý

```go
allowUnauthenticated: map[string]bool{
    protocol.MethodInitialize: true,
},
```

Field này được khai báo trong `NewAuthMiddleware` nhưng... **không được dùng** trong `Handler` hay `OptionalHandler` ở file này. Có 2 khả năng:

- Logic check method này nằm ở chỗ khác (ví dụ một middleware layer bên ngoài)
- Hoặc đây là phần chưa implement xong — `allowUnauthenticated` được chuẩn bị sẵn cho tương lai