package main

import (
	"fmt"
	"sync"
	"time"
)

// 2. Once: Giải quyết bài toán "Singleton Connection"
// Mục đích: Tránh việc tạo hàng ngàn kết nối thừa thãi làm sập Database.

type DatabaseConnection struct {
	connectionString string
}

var db *DatabaseConnection
var once sync.Once

func GetDatabaseConnection(connectionString string) *DatabaseConnection {
	once.Do(func() {
		db = &DatabaseConnection{connectionString: connectionString}
		fmt.Println("Đã tạo kết nối Database thành công!")
	})
	return db
}

func main() {
	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn := GetDatabaseConnection("user:pass@tcp(localhost:3306)/mydb")
			fmt.Printf("Người dùng %d: Đã nhận được kết nối\n", id)
		}(i)
	}

	wg.Wait()
}
