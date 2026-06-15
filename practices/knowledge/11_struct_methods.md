Chào bạn, kiến thức về **Structs** và **Methods** trong Go là nền tảng cực kỳ quan trọng để lập trình hướng đối tượng (theo phong cách Go). Dưới đây là phần trình bày lại một cách cô đọng, có hệ thống và dễ hiểu nhất:

---

## 1. Structs (Cấu trúc dữ liệu)

Struct là một kiểu dữ liệu do người dùng tự định nghĩa, cho phép nhóm các trường dữ liệu (fields) có liên quan lại với nhau. Hãy nghĩ về nó như một "bản thiết kế" cho các đối tượng.

### Định nghĩa và Khởi tạo
* **Khai báo:** Sử dụng từ khóa `type` và `struct`.
* **Zero value:** Nếu không khởi tạo giá trị, các trường sẽ nhận giá trị mặc định (`0`, `""`, `false`).
* **Struct Literal:** Khởi tạo nhanh bằng cách liệt kê giá trị.



```go
type Person struct {
    Name string
    Age  int
}

// Khởi tạo có tên trường (khuyên dùng)
p := Person{Name: "Karan", Age: 22}

// Khởi tạo không tên trường (phải đúng thứ tự và đủ số lượng)
p2 := Person{"Bruce", 40}

// Anonymous struct (Dùng một lần, không cần đặt tên kiểu)
a := struct{ Title string }{"Golang"}
```

### Đặc điểm quan trọng
* **Value Type:** Struct là kiểu tham trị. Khi gán `p2 = p1`, Go sẽ tạo một bản sao hoàn toàn mới. Thay đổi trên `p2` không ảnh hưởng đến `p1`.
* **Exported (Public/Private):** Nếu tên trường viết hoa (ví dụ `Name`), nó có thể được truy cập từ package khác. Nếu viết thường (`age`), nó là nội bộ.
* **Pointer to Struct:** Go hỗ trợ "Syntactic Sugar" giúp bạn truy cập trường qua con trỏ mà không cần giải tham chiếu thủ công.
    ```go
    ptr := &p
    fmt.Println(ptr.Name) // Thay vì (*ptr).Name
    ```
* **Empty Struct:** `struct{}` chiếm **0 byte** bộ nhớ, thường dùng trong các bài toán tối ưu hoặc làm tín hiệu trong Channel.

---

## 2. Embedding & Composition (Nhúng và Hợp thành)

Go không có kế thừa (inheritance) như Java hay C++, thay vào đó Go dùng **Composition**.

* **Embedding (Nhúng):** Cho phép một struct "thừa hưởng" các trường và phương thức của struct khác bằng cách khai báo tên kiểu mà không có tên trường.
    ```go
    type SecretAgent struct {
        Person // Embedding
        LicenseToKill bool
    }
    ```
* **Composition (Hợp thành):** Khai báo một struct như một trường bình thường. Đây là cách được khuyến khích hơn để giữ code tường minh.

---

## 3. Methods (Phương thức)

Method bản chất là một hàm, nhưng có thêm một đối số đặc biệt gọi là **Receiver** (bộ nhận).

### Cấu trúc một Method
```go
func (receiver Tên_Biến Tên_Kiểu) Tên_Hàm(tham_số) trả_về { ... }
```

### Phân loại Receiver
Có hai loại Receiver quan trọng mà bạn cần phân biệt rõ:

| Đặc điểm | **Value Receiver** `(p Person)` | **Pointer Receiver** `(p *Person)` |
| :--- | :--- | :--- |
| **Cơ chế** | Làm việc trên một **bản sao**. | Làm việc trên **địa chỉ gốc**. |
| **Thay đổi** | Không ảnh hưởng đến biến gốc. | **Có thay đổi** được biến gốc. |
| **Hiệu năng** | Tốn bộ nhớ nếu struct lớn (do copy). | Hiệu năng cao (chỉ truyền địa chỉ). |



### Tại sao dùng Method thay vì Function?
1.  **Tính đóng gói:** Giúp code gọn gàng, giống phong cách hướng đối tượng (`object.Method()`).
2.  **Tránh xung đột tên:** Nhiều kiểu dữ liệu khác nhau có thể có cùng tên phương thức (ví dụ: cả `Circle` và `Square` đều có method `Area()`).
3.  **Interface:** Đây là điều kiện tiên quyết để một struct có thể "thỏa mãn" một Interface (mình sẽ học ở phần sau).

---

### Ví dụ tổng hợp:
```go
type Rect struct {
    Width, Height float64
}

// Value receiver: chỉ tính toán, không sửa đổi
func (r Rect) Area() float64 {
    return r.Width * r.Height
}

// Pointer receiver: có thể sửa đổi giá trị bên trong
func (r *Rect) Scale(f float64) {
    r.Width *= f
    r.Height *= f
}
```
