package main

import (
	"context"
	"fmt"
	"time"
)

// khởi tạo context: 
// context.Background(): Context gốc, không có giá trị, không có timeout.
// context.WithCancel(parent): Tạo ra một context mới từ context cha và một hàm cancel.
// context.WithTimeout(parent, duration): Tạo ra một context mới từ context cha với timeout.
// context.WithDeadline(parent, time): Tạo ra một context mới từ context cha với deadline.
// context.WithValue(parent, key, value): Tạo ra một context mới từ context cha với value.


// 1. Tham số đầu tiên: Luôn để ctx context.Context là đối số đầu tiên của hàm.
// 2. Không lưu vào Struct: Context chỉ nên được truyền qua các hàm, không nên lưu nó vào trong một thuộc tính của struct.
// 3. Gọi cancel(): Bất kể bạn dùng WithCancel hay WithTimeout, luôn phải gọi hàm cancel() (thường dùng defer cancel()) để tránh rò rỉ bộ nhớ (memory leak).
// 4. Immutable: Context không thể thay đổi giá trị bên trong. Khi bạn thêm Timeout hay Value, Go thực chất tạo ra một bản sao mới dựa trên context cũ.


func fetchWeather(ctx context.Context) {
	select {
	case <-time.After(5 * time.Second):
		fmt.Println("Lấy dữ liệu thành công")
	case <-ctx.Done():
		fmt.Println("Lỗi:", ctx.Err())
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go fetchWeather(ctx)
	time.Sleep(3 * time.Second)
}
