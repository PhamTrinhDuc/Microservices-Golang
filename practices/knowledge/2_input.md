### 1. Sử dụng `fmt.Scan` (Cách đơn giản nhất)
Đây là cách nhanh nhất để lấy một giá trị cơ bản. `fmt.Scan` sẽ tự động tìm các giá trị phân tách bằng khoảng trắng.

```go
package main

import "fmt"

func main() {
    var n int
    fmt.Print("Nhập số N: ")
    fmt.Scan(&n) // Truyền địa chỉ của biến n
    fmt.Printf("Số bạn vừa nhập là: %d\n", n)
}
```

### 2. Sử dụng `fmt.Scanf` (Khi cần định dạng cụ thể)
Nếu bạn muốn người dùng nhập theo một khuôn mẫu nhất định (ví dụ: `ID: 100`), bạn dùng `Scanf`.

```go
fmt.Scanf("%d", &n)
```

### 3. Sử dụng `bufio` và `strconv` (Cách chuyên nghiệp)
Cách này thường được dùng trong các bài tập thuật toán hoặc ứng dụng thực tế vì nó **nhanh hơn** và xử lý chuỗi tốt hơn khi người dùng vô tình nhập ký tự lạ.

```go
package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)

func main() {
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Nhập số N: ")

    // Đọc dữ liệu đến khi gặp dòng mới
    input, _ := reader.ReadString('\n')

    // Loại bỏ khoảng trắng/ký tự xuống dòng và chuyển thành số
    input = strings.TrimSpace(input)
    n, err := strconv.Atoi(input)

    if err != nil {
        fmt.Println("Lỗi: Vui lòng nhập một số nguyên hợp lệ!")
    } else {
        fmt.Printf("Số N của bạn là: %d\n", n)
    }
}
```

---

### Tóm tắt sự khác biệt

| Đặc điểm | `fmt.Scan` | `bufio` + `strconv` |
| :--- | :--- | :--- |
| **Độ khó** | Rất dễ | Trung bình |
| **Tốc độ** | Chậm hơn | Rất nhanh |
| **Xử lý lỗi** | Hạn chế | Rất chi tiết |
| **Phù hợp** | Viết code nhanh, test nhỏ | Dự án thực tế, Competitive Programming |

**Một lưu ý nhỏ:** Trong Go, khi dùng các hàm `Scan`, bạn luôn phải truyền **con trỏ** (dấu `&` trước tên biến) để hàm có thể ghi trực tiếp giá trị vào ô nhớ của biến đó.
