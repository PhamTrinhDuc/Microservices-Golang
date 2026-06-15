package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func generator(done <-chan struct{}, orders []int) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)
		for _, value := range orders {
			select {
			case out <- value: // đẩy id vào channels
			case <-done: // nếu nhận tín hiệu dừng -> thoát
				return
			}
		}
	}()
	return out
}

func checkStock(done <-chan struct{}, orders <-chan int) <-chan string {
	out := make(chan string)

	go func() {
		defer close(out)
		for id := range orders {
			time.Sleep(500 * time.Millisecond)
			result := fmt.Sprintf("Đơn hàng #%d: Đã kiểm kho", id)

			select {
			case out <- result: // đẩy kết quả vào channel
				// case <-done: // nếu nhận tín hiệu dừng -> thoát
				// 	return
			}
		}
	}()
	return out
}

// ================================ GOROUTINES ==================================
func goroutines() {
	orders := []int{101, 102, 103, 104, 105}
	done := make(chan struct{})

	stage1 := generator(done, orders)
	for i := 0; i < 3; i++ {
		fmt.Println(<-stage1)
	}

	// time.Sleep(time.Second) // Đợi 1 chút để đảm bảo các tiến trình đã chạy
	// fmt.Println("Số Goroutine đang bị kẹt trước khi close:", runtime.NumGoroutine())

	// close(done)

	// time.Sleep(time.Second) // Đợi 1 chút để đảm bảo các tiến trình đã chạy
	// fmt.Println("Số Goroutine đang bị kẹt sau khi close:", runtime.NumGoroutine())

	stage2 := checkStock(done, stage1)
	for i := 0; i < 3; i++ {
		fmt.Println(<-stage2)
	}

	time.Sleep(time.Second) // Đợi 1 chút để đảm bảo các tiến trình đã chạy
	fmt.Println("Số Goroutine đang bị kẹt trước khi close:", runtime.NumGoroutine())

	close(done)

	time.Sleep(time.Second) // Đợi 1 chút để đảm bảo các tiến trình đã chạy
	fmt.Println("Số Goroutine đang bị kẹt sau khi close:", runtime.NumGoroutine())

}

// ===================================== DEADLOCK =================================
func deadlock() {
	var c chan string
	c <- "Hello"
	fmt.Println(c)
}

// ===================================== WAITGROUP ==============================
func worker(id int, wg *sync.WaitGroup) {
	// 2. Đảm bảo luôn gọi Done khi thoát hàm
	defer wg.Done()

	fmt.Printf("Công nhân %d: Đang bắt đầu làm việc...\n", id)
	time.Sleep(time.Second) // Giả lập làm việc nặng
	fmt.Printf("Công nhân %d: ĐÃ XONG!\n", id)
}

func waitGroup() {
	var wg sync.WaitGroup

	// for i := 1; i <= 3; i++ {
	// 	// 1. Đăng ký: Thêm 1 người vào danh sách chờ
	// 	wg.Add(1)
	// 	go worker(i, &wg) // Lưu ý: Truyền con trỏ &wg
	// }

	// fmt.Println("Main: Đang đứng đợi các công nhân...")

	// // 3. Chặn ở đây, không cho main thoát
	// wg.Wait() // không có wait, 3 goroutines + main sẽ chờ

	// fmt.Println("Main: Tất cả đã xong, đóng cửa đi về!")
	// fmt.Println("Số Goroutine đang bị kẹt trước khi close:", runtime.NumGoroutine())

	wg.Add(4)
	go worker(1, &wg)
	go worker(2, &wg)
	go worker(3, &wg)
	// deadlock vì waitGroup add 4 nhưng worker 4 không làm việc và main cũng đang đợi => không có ai làm việc => deadlock
	// wg.Wait()
	fmt.Println("Số Goroutine đang bị kẹt sau khi close:", runtime.NumGoroutine())
}

func main() {
	waitGroup()
}
