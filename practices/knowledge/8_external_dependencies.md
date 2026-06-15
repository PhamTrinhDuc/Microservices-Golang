Phần **External Dependencies** (Thư viện bên ngoài) chính là lúc bạn tận dụng sức mạnh của cộng đồng Go toàn cầu. Thay vì tự viết mọi thứ từ đầu (như logger, kết nối database, hay mã hóa), bạn chỉ cần "mượn" các thư viện xịn xò đã có sẵn.

Có 3 điểm "mấu chốt" bạn cần lưu ý ở phần này:

---

### 1. `go install` vs `go get` vs `go mod tidy` (Sự khác biệt quan trọng)

Trong tài liệu bạn gửi có nhắc đến `go install`, nhưng thực tế khi làm dự án, bạn sẽ dùng bộ 3 này như sau:

* **`go get <url>`**: Dùng khi bạn muốn **tải một thư viện cụ thể** về và ghi nó vào file `go.mod`.
    * Ví dụ: `go get github.com/rs/zerolog`
* **`go install <url>`**: Thường dùng để cài đặt các **công cụ (tools)** chạy bằng dòng lệnh (CLI). Nó sẽ tải về, biên dịch và ném file `.exe` vào thư mục `$GOPATH/bin`. 
* **`go mod tidy` (Khuyên dùng nhất)**: Bạn cứ vào code gõ `import "github.com/..."`, sau đó ra Terminal gõ `go mod tidy`. Go sẽ tự hiểu bạn đang thiếu thư viện đó và tự đi tải về phiên bản phù hợp nhất.

---

### 2. Cách đọc Tài liệu (Go Doc)
Go có một văn hóa rất hay là: **Code chính là tài liệu**.
* Các lập trình viên Go thường viết chú thích (comment) ngay trên hàm/biến.
* Công cụ `go doc` sẽ quét các chú thích đó để tạo ra trang web hướng dẫn cực kỳ chuyên nghiệp.
* **Mẹo cho bạn:** Để xem tài liệu của bất kỳ thư viện nào, bạn chỉ cần vào trang: `pkg.go.dev`. Ví dụ: `pkg.go.dev/github.com/rs/zerolog`.

---

### 3. Quy tắc "Không có quy tắc" (Folder Structure)
Như tài liệu bạn gửi nói: **Go không bắt ép bạn phải tổ chức thư mục theo một khuôn mẫu cứng nhắc** (như Java hay Ruby on Rails). 

Tuy nhiên, trong thực tế, người ta thường theo các cấu trúc phổ biến như:
* `/cmd`: Chứa các file `main.go` (điểm bắt đầu của ứng dụng).
* `/internal`: Chứa code quan trọng mà bạn **không muốn** các dự án khác bên ngoài được phép import.
* `/pkg`: Chứa code mà bạn muốn chia sẻ cho mọi người dùng chung.

---

### Một ví dụ thực tế khi dùng thư viện `zerolog`:

Tại sao người ta lại dùng `zerolog` thay vì `fmt.Println`? 
Vì khi làm Backend chuyên nghiệp, bạn cần log có định dạng JSON để các hệ thống theo dõi (như ELK, Grafana) có thể đọc được:

```go
package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
    // Cấu hình log theo kiểu Unix (số giây)
    zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

    log.Info().
        Str("user", "PhamTrinhDuc").
        Int("attempt", 1).
        Msg("Người dùng đang thử đăng nhập")
}
```
**Kết quả in ra sẽ cực kỳ chuyên nghiệp:**
`{"level":"info","user":"PhamTrinhDuc","attempt":1,"time":1711360000,"message":"Người dùng đang thử đăng nhập"}`

---

### Tổng kết phần này:
1.  **Import:** Dùng URL đầy đủ của GitHub (hoặc GitLab công ty).
2.  **Quản lý:** File `go.mod` sẽ tự động cập nhật khi bạn chạy `go mod tidy`.
3.  **Tổ chức:** Cứ làm sao cho nó "đơn giản và dễ hiểu" (simple and intuitive) là được, đừng làm phức tạp hóa vấn đề.