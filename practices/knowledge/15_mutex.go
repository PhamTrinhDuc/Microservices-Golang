package main

import (
	"fmt"
	"sync"
	"time"
)

// Mutex: Giải quyết bài toán "Rút tiền ATM"
// Mục đích: Đảm bảo tính nhất quán dữ liệu khi nhiều nơi cùng sửa một con số quan trọng.

type BankAccount struct {
	sync.Mutex
	balance int
}

func (a *BankAccount) WithDraw(amount int, wg *sync.WaitGroup) {
	defer wg.Done()
	defer a.Unlock()

	a.Lock()
	if a.balance >= amount {
		time.Sleep(10 * time.Millisecond)
		a.balance -= amount
		fmt.Printf("Rút %d thành công. Còn lại: %d\n", amount, a.balance)
	} else {
		fmt.Printf("Rút %d thất bại. Không đủ tiền!\n", amount)
	}
}

func main() {
	var wg sync.WaitGroup

	account := BankAccount{balance: 100}

	wg.Add(3)
	go account.WithDraw(30, &wg)
	go account.WithDraw(50, &wg)
	go account.WithDraw(40, &wg)

	wg.Wait()

	fmt.Printf("Số dư cuối cùng: %d\n", account.balance)
}

// Giả sử số dư đang là 100. Hai Goroutine (G1 rút 80, G2 rút 50) cùng chạy:
// 1. G1 kiểm tra: if 100 >= 80 (Đúng).
// 2. G1 chuẩn bị trừ tiền, nhưng lúc này time.Sleep bắt nó dừng lại một chút.
// 3. G2 tranh thủ lúc G1 đang ngủ, nhảy vào kiểm tra: if 100 >= 50 (Đúng - vì G1 đã trừ đâu!).
// 4. G1 thức dậy, thực hiện: balance = 100 - 80 (Còn 20).
// 5. G2 cũng thức dậy, thực hiện: balance = 20 - 50 (Còn -30).
