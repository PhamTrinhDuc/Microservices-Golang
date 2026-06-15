Phần **Package** này chính là cách Go tổ chức code để tái sử dụng và quản lý quyền truy cập (Private/Public). Nếu Module là cái "thùng" lớn chứa cả dự án, thì Package là những cái "hộp" nhỏ bên trong.

Có 3 quy tắc "sống còn" về Package trong Go mà bạn cần nhớ:

---

### 1. Quy tắc Viết hoa - Viết thường (Export vs Unexport)
Đây là điểm cực kỳ thông minh của Go: **Không cần từ khóa `public` hay `private` như Java/C#.**
* **Viết hoa chữ cái đầu:** Là **Public** (Exported). Các package khác có thể gọi được.
    * Ví dụ: `func SayHello()` hoặc `var Version`.
* **Viết thường chữ cái đầu:** Là **Private** (Unexported). Chỉ những file nằm trong **cùng một thư mục** (cùng package) mới thấy nhau.
    * Ví dụ: `func calculate()` hoặc `var secretKey`.

> **Mẹo:** Nếu bạn viết `custom.value` mà bị báo lỗi, hãy kiểm tra ngay xem đã đổi thành `custom.Value` (viết hoa chữ V) chưa nhé.

---

### 2. Package `main` và Hàm `main()`
* Mọi file Go đều phải khai báo `package <tên>` ở dòng đầu tiên.
* **Package `main`:** Đây là package đặc biệt. Go chỉ tạo ra file chạy (`.exe`) khi nó thấy package `main`.
* **Hàm `func main()`:** Là "cửa ngõ" (entry point). Khi bạn chạy chương trình, Go sẽ tìm đến đây đầu tiên để thực thi.
* Các package khác (không phải `main`) thường được gọi là **Library** (Thư viện) — chúng chỉ nằm đó chờ được package khác gọi tới.

---

### 3. Cách Import từ Module của mình
Đây là chỗ người mới hay bị rối nhất. Khi bạn muốn dùng package `custom` trong `main.go`, bạn phải dựa vào tên module đã khai báo trong file `go.mod`.

Giả sử file `go.mod` của bạn là: `module example`
Cấu trúc thư mục:
```text
Learn-Go/
├── go.mod
├── main.go
└── custom/
    └── code.go
```

Trong `main.go`, bạn phải gọi:
```go
import "example/custom" // Tên_module / Tên_thư_mục_package
```

> [!IMPORTANT]
> **Lưu ý cực kỳ quan trọng:**
> * **Không bao giờ được comment dòng `package ...`**: Nếu bạn comment dòng này, file sẽ không thuộc về package nào và Go sẽ không thể import được.
> * **Hàm phải tồn tại:** Nếu bạn comment hết nội dung hoặc cả định nghĩa hàm (`func`), các file khác gọi đến sẽ bị lỗi `undefined`.
> * **Thư mục không được rỗng:** Nếu tất cả file Go trong một thư mục bị comment sạch code, Go sẽ báo lỗi "no Go files in..." khi bạn cố gắng import thư mục đó.

---

### 4. Alias (Đặt tên biệt hiệu)
Khi bạn import hai package có tên giống hệt nhau (ví dụ cả hai đều tên là `network`), bạn dùng Alias để tránh xung đột:
```go
import (
    net1 "example/network"
    net2 "google/network"
)
```

---

### Một ví dụ "thực chiến" cho bạn:
Hãy tưởng tượng bạn đang viết code cho công ty VTI:
1. Bạn tạo package `auth` để xử lý đăng nhập.
2. Bạn để hàm `checkPassword` viết thường (vì không muốn ai ngoài package này đụng vào).
3. Bạn để hàm `Login` viết hoa để các package khác gọi dùng.

**Câu hỏi nhỏ cho bạn:** Nếu trong file `code.go` thuộc package `custom`, bạn khai báo `var age int = 30`. Sang file `main.go`, bạn gõ `fmt.Println(custom.age)` thì code có chạy được không? 

(Gợi ý: Hãy nhìn vào chữ cái đầu của `age` nhé!)