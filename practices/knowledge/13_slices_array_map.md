
## 1. ARRAY (Mảng) — Hộp cố định

Hãy tưởng tượng array như một **dãy ngăn tủ có số lượng cố định**. Bạn khai báo 4 ngăn thì mãi mãi chỉ có 4 ngăn.

### Khai báo & Khởi tạo

```go
// Khai báo — tự động điền 0 vào hết
var arr [4]int
// → [0 0 0 0]

// Khai báo + gán giá trị ngay
arr := [4]int{1, 2, 3, 4}
// → [1 2 3 4]
```

### Truy cập phần tử

```go
arr := [4]int{10, 20, 30, 40}
fmt.Println(arr[0]) // → 10  (index bắt đầu từ 0)
fmt.Println(arr[3]) // → 40
```

### Duyệt mảng (2 cách phổ biến)

```go
arr := [4]int{1, 2, 3, 4}

// Cách 1: dùng for + len
for i := 0; i < len(arr); i++ {
    fmt.Println(i, arr[i])
}

// Cách 2: dùng range (gọn hơn, Go-style)
for i, e := range arr {
    fmt.Println(i, e)
}
```

> 💡 **range** trả về 2 giá trị: **index** và **element**. Nếu không cần index, dùng `_` để bỏ qua: `for _, e := range arr`

### ⚠️ Điểm quan trọng cần nhớ về Array

```go
a := [4]int{1, 2, 3, 4}
b := a         // b là BẢN SAO của a
b[0] = 999

fmt.Println(a) // → [1 2 3 4]  — a KHÔNG thay đổi
fmt.Println(b) // → [999 2 3 4]
```

> Array trong Go là **value type** — gán = sao chép toàn bộ. Khác với C/Java!

---

## 2. SLICE — Mảng linh hoạt ⭐

Slice là thứ bạn sẽ dùng **90% thời gian** thay vì array. Nó giống array nhưng **không cố định kích thước**.

### Slice gồm 3 thứ bên trong:
```
┌─────────────────────────────────┐
│  Pointer → trỏ vào array gốc   │
│  Length  → số phần tử hiện có  │
│  Capacity → có thể chứa tối đa │
└─────────────────────────────────┘
```

### Khai báo & Khởi tạo

```go
// Khai báo — zero value của slice là nil (khác array!)
var s []string
fmt.Println(s == nil) // → true

// Cách 1: dùng make(type, length, capacity)
s := make([]string, 0, 5)

// Cách 2: slice literal (dùng nhiều nhất)
s := []string{"Go", "Python", "Rust"}

// Cách 3: cắt từ array
a := [5]int{10, 20, 30, 40, 50}
s := a[1:4]  // lấy index 1, 2, 3  → [20 30 40]
s := a[:3]   // lấy từ đầu đến 2   → [10 20 30]
s := a[2:]   // lấy từ 2 đến cuối  → [30 40 50]
```

### Thêm phần tử — `append`

```go
s := []string{"a", "b", "c"}
s = append(s, "d", "e")
fmt.Println(s) // → [a b c d e]
```

> 💡 Luôn gán lại: `s = append(s, ...)` — vì append trả về slice mới

### Sao chép slice — `copy`

```go
src := []int{1, 2, 3, 4}
dst := make([]int, len(src))
copy(dst, src)

dst[0] = 999
fmt.Println(src) // → [1 2 3 4]  — src không bị ảnh hưởng
fmt.Println(dst) // → [999 2 3 4]
```

### ⚠️ Điểm quan trọng về Slice

```go
a := [5]string{"Mon", "Tue", "Wed", "Thu", "Fri"}
s := a[0:2]   // slice trỏ vào array a

s[0] = "Sunday"

fmt.Println(a) // → [Sunday Tue Wed Thu Fri]  — a bị thay đổi theo!
fmt.Println(s) // → [Sunday Tue]
```

> Slice là **reference type** — nó trỏ vào array gốc, sửa slice = sửa luôn array!

---

## 3. MAP — Từ điển key-value

Map giống như một **cuốn từ điển**: tra bằng **key**, nhận về **value**.

### Khai báo & Khởi tạo

```go
// Dùng make
m := make(map[string]int)

// Dùng map literal
m := map[string]int{
    "apple":  5,
    "banana": 3,
}
```

### Các thao tác cơ bản

```go
m := map[string]int{"a": 1, "b": 2}

// Thêm / Cập nhật
m["c"] = 3
m["a"] = 99  // ghi đè

// Đọc
fmt.Println(m["c"])  // → 3

// Kiểm tra key có tồn tại không
val, ok := m["c"]
if ok {
    fmt.Println("Có key c, giá trị:", val)
} else {
    fmt.Println("Không có key c")
}

// Xóa
delete(m, "b")

// Duyệt
for key, value := range m {
    fmt.Println(key, value)
}
```

> ⚠️ Thứ tự duyệt map **không cố định** — mỗi lần chạy có thể khác nhau!

### Map cũng là reference type

```go
m1 := map[string]int{"a": 1}
m2 := m1       // m2 và m1 cùng trỏ vào 1 chỗ

m2["b"] = 2

fmt.Println(m1) // → map[a:1 b:2]  — m1 cũng thay đổi!
```

---

## 📋 Bảng tóm tắt nhanh

| | Array | Slice | Map |
|---|---|---|---|
| Kích thước | Cố định | Linh hoạt | Linh hoạt |
| Zero value | `[0 0 0]` | `nil` | `nil` |
| Type | Value | Reference | Reference |
| Dùng khi | Ít gặp | Thường xuyên | Key-value |

---