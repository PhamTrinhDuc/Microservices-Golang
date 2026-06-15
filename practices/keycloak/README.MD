# Keycloak dev playground

Mục đích: chạy thử Keycloak ở local, tạo realm/client/user, lấy 1 JWT,
và verify token đó bằng Go — trước khi đụng vào migrate hệ thống thật.

## 1. Chạy Keycloak

```bash
docker compose up -d
```

Đợi vài chục giây, sau đó vào http://localhost:8080 — login admin console
bằng `admin` / `admin`.

Lưu ý: `start-dev` là mode dev, không dùng cho production (không TLS,
một số check bảo mật bị tắt).

## 2. Tạo realm

Trong Admin Console: góc trên trái → "Create realm" → đặt tên `demo`.

## 3. Tạo client

Vào realm `demo` → Clients → Create client:

- Client ID: `demo-client`
- Client authentication: **Off** (public client, đơn giản cho dev)
- Trong tab "Capability config" (sau khi tạo): bật **Direct access grants**
  — cần thiết để lấy token qua `curl` bằng username/password (password grant).
  Chỉ dùng cho test, không nên bật ở production.

## 4. Tạo user

Realm `demo` → Users → Add user:

- Username: `testuser`
- Tắt "Email verified" nếu không cần, hoặc bật cho gọn

Sau khi tạo, vào tab "Credentials" → Set password:

- Đặt password, ví dụ `testpass`
- "Temporary": **Off** (nếu để On, lần đầu login sẽ bị bắt đổi password)

## 5. Lấy access token

```bash
curl -s -X POST http://localhost:8080/realms/demo/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=password" \
  -d "client_id=demo-client" \
  -d "username=testuser" \
  -d "password=testpass" | jq
```

Response trả về `access_token`, `refresh_token`, `id_token`. Copy giá trị
`access_token` (chuỗi rất dài) để dùng ở bước sau.

Bạn có thể tự decode thử ở https://jwt.io để xem claims bên trong
(chỉ là decode, không phải verify).

## 6. Verify token bằng Go

```bash
cd verify-token
go mod tidy
go run main.go "<access_token>"
```

Nếu hợp lệ, chương trình in ra toàn bộ claims (sub, email, exp, iss, aud...).
Nếu sửa 1 ký tự trong token rồi chạy lại, sẽ thấy lỗi "token không hợp lệ" —
đây chính là cơ chế verify bằng public key (JWKS) đã nói ở các phần trước.

## 7. Vài thứ nên thử thêm

- Mở http://localhost:8080/realms/demo/.well-known/openid-configuration
  — đây là endpoint discovery, chứa `jwks_uri`, `authorization_endpoint`,
  `token_endpoint`... Mọi service (Go) sẽ dựa vào endpoint này để biết
  lấy public key ở đâu.

- Mở `jwks_uri` (lấy URL từ trên) — đây chính là public key dạng JSON,
  tương đương file `public_key.pem` trong code demo trước, nhưng do
  Keycloak tự quản lý và có thể rotate.

- Thử tạo thêm 1 realm role (Realm roles → Create role, ví dụ `admin`),
  gán cho `testuser` (Users → testuser → Role mapping), rồi vào
  Client scopes → `demo-client-dedicated` → Mappers → thêm mapper
  "User Realm Role" để role này xuất hiện trong access token dưới
  dạng claim — đây là cách thay thế cho field `Role` trong
  `JWTClaims` của bạn.

- Thử login với Google: Identity providers → Add provider → Google,
  điền Client ID/Secret từ Google Cloud Console. Sau đó trang login
  của realm `demo` sẽ có thêm nút "Sign in with Google".

## Dọn dẹp

```bash
docker compose down -v
```

`-v` xóa luôn volume database — dùng khi muốn làm lại từ đầu sạch sẽ.