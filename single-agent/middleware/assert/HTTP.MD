## 1. `http.Handler` và `http.HandlerFunc` là gì

`http.Handler` là một **interface** của Go:

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

Tức là bất kỳ struct nào có method `ServeHTTP` thì đều được coi là một `Handler`. Go's HTTP server chỉ biết gọi đúng một thứ: `ServeHTTP`.

Vậy `http.HandlerFunc` là gì? Đây là trick của Go stdlib:

```go
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
    f(w, r)
}
```

Nó **wrap một function bình thường thành một Handler**. Nhờ vậy thay vì phải tạo struct, bạn chỉ cần viết:

```go
http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // logic ở đây
})
```

Viết logic trong một cái hàm có chữ ký `(w http.ResponseWriter, r *http.Request)` và đưa nó vào trong `http.HandlerFunc `

> WHY? Tại sao phải viết 2 hàm `Handler` và `OptionalHandler`? 
Bởi vì bộ Router hoặc Middleware chỉ nhận các hàm logic có dạng `http.Handler`

---

## 2. `w` và `r` dùng để làm gì

```
HTTP Request
┌──────────────────────────────────┐
│  r *http.Request                 │  ← đọc từ client
│  ├─ r.Header.Get("Authorization")│
│  ├─ r.Context()                  │
│  ├─ r.URL, r.Method, r.Body      │
└──────────────────────────────────┘

┌──────────────────────────────────┐
│  w http.ResponseWriter           │  ← ghi về cho client
│  ├─ w.Header().Set("Content-Type"│
│  ├─ w.WriteHeader(401)           │
│  └─ w.Write([]byte(...))         │
└──────────────────────────────────┘
```

`r` là **read-only** — dữ liệu đến từ client. `w` là **write-only** — bạn dùng nó để trả response. Trong `sendError`:

```go
func (m *AuthMiddleware) sendError(w http.ResponseWriter, ...) {
    w.Header().Set("Content-Type", "application/json")  // set header
    w.WriteHeader(http.StatusUnauthorized)               // set status 401
    json.NewEncoder(w).Encode(response)                  // ghi body
}
```

`w` hoạt động như một `io.Writer` — `json.NewEncoder(w)` ghi thẳng JSON vào response stream.

---

## 3. `r.WithContext(ctx)` — tại sao phải clone

`http.Request` là một **struct**, không phải con trỏ đến map hay slice — nên muốn thay đổi an toàn phải tạo bản sao:

```go
// KHÔNG làm thế này — sửa trực tiếp r có thể ảnh hưởng goroutine khác
r.ctx = newCtx  // không được, ctx là unexported field

// Đúng cách — WithContext trả về bản sao mới của r với ctx mới
ctx := auth.WithAuth(r.Context(), claims)
r2 := r.WithContext(ctx)         // r2 là struct mới, r cũ không đổi
next.ServeHTTP(w, r2)
```

Thực ra trong code viết gọn luôn:

```go
next.ServeHTTP(w, r.WithContext(ctx))
```

Context chạy xuyên suốt từ middleware → handler → database. Đây chính là con đường `TenantID` đi từ JWT token xuống tới `SET LOCAL` trong PostgreSQL.

---

## 4. `next.ServeHTTP(w, r)` — cơ chế middleware chain

Đây là phần hay nhất. Hình dung middleware như **vòng bọc hành tây**:

```
Request ──►  Middleware A
                │  (xử lý trước)
                ▼
             Middleware B
                │  (xử lý trước)
                ▼
             Handler thật
                │
                ▼
             Middleware B  (xử lý sau nếu có)
                │
                ▼
             Middleware A  (xử lý sau nếu có)
                │
Response ◄──────┘
```

`next` là **Handler tiếp theo trong chain** — được truyền vào qua tham số:

```go
func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

        // === PHẦN TRƯỚC ===
        claims, err := m.validator.ValidateToken(authHeader)
        if err != nil {
            m.sendError(...)
            return          // ← chặn chain tại đây nếu lỗi
        }

        ctx := auth.WithAuth(r.Context(), claims)

        // === GỌI TIẾP THEO ===
        next.ServeHTTP(w, r.WithContext(ctx))  // ← trao quyền cho handler sau

        // === PHẦN SAU (nếu cần) ===
        // có thể làm gì đó sau khi handler chạy xong, vd: logging
    })
}
```

Khi gọi `return` trước `next.ServeHTTP` — chain bị **dừng hoàn toàn**, handler thật không bao giờ chạy. Đây là cách middleware chặn request không hợp lệ.