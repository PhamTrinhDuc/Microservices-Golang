# 🧪 Golang Testing Cheat Sheet

Hướng dẫn nhanh cách viết và chạy Unit Test trong dự án Go.

## 1. Quy tắc đặt tên
* **File test:** Phải kết thúc bằng đuôi `_test.go`. (VD: `db_test.go`).
* **Hàm test:** Phải bắt đầu bằng chữ `Test` và nhận tham số `(t *testing.T)`. (VD: `func TestNewDB(t *testing.T)`).
* **Package:** Thường dùng cùng tên package với file gốc hoặc thêm đuôi `_test` để đảm bảo tính đóng gói (VD: `package database_test`).

## 2. Các lệnh chạy Test thông dụng

| Lệnh | Ý nghĩa |
| :--- | :--- |
| `go test ./...` | Chạy toàn bộ test trong dự án. |
| `go test -v ./...` | **(Khuyên dùng)** Chạy test và hiển thị log chi tiết (Verbose). |
| `go test -run TestName` | Chỉ chạy một hàm test cụ thể. |
| `go test -cover` | Kiểm tra tỉ lệ bao phủ code (Coverage). |

## 3. Cấu trúc Table-Driven Test (Khuyên dùng)
Đây là cách viết test chuẩn nhất trong Go, giúp quản lý nhiều bộ dữ liệu (test cases) trên cùng một logic.

```go
func TestFunctionName(t *testing.T) {
    // 1. Định nghĩa danh sách test cases
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"hợp lệ", "data", false},
        {"lỗi rỗng", "", true},
    }

    for _, tc := range tests {
        // 2. Chạy sub-test cho từng case
        t.Run(tc.name, func(t *testing.T) {
            err := YourFunction(tc.input)
            
            // 3. Kiểm tra kết quả
            if tc.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## 4. Các lưu ý quan trọng

### Tại sao log không in ra?
* Mặc định Go test sẽ **nuốt (giấu)** các dòng `t.Log` hoặc `fmt.Print` nếu test đó **Pass**.
* **Giải pháp:** Luôn thêm cờ `-v` khi chạy test: `go test -v`.

### Assertions (Thư viện hỗ trợ)
Thay vì dùng `if err != nil { t.Errorf(...) }` thủ công, nên dùng thư viện `testify` để code sạch hơn:
```bash
go get github.com/stretchr/testify
```
* `assert.Equal(t, expected, actual)`
* `assert.NotNil(t, object)`
* `assert.Error(t, err)`

### Kiểm tra Coverage (Độ bao phủ)
Để xem những dòng code nào chưa được test quét qua:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # Mở giao diện web xem trực quan
```

