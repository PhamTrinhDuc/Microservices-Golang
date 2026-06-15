package main

import (
	"fmt"
)

func array_func() {
	// 1. // Khai báo — tự động điền 0 vào hết
	var arr [4]int
	fmt.Println(arr)
	// 2. Khai báo + gán giá trị ngay
	var arr2 = [4]int{1, 2, 3, 4}
	fmt.Println(arr2)

	// 3. Truy cập phần tử
	arr3 := [4]int{10, 20, 30, 40}
	fmt.Println(arr3[0]) // 10
	fmt.Println(arr3[2]) // 30

	// 4. Độ dài mảng
	fmt.Println(len(arr3)) // 4

	// 5. Duyệt mảng
	arr4 := [4]int{5, 10, 15, 20}
	for i := 0; i < len(arr4); i++ {
		fmt.Println(arr4[i])
	}

	// using range to loop through array
	for index, value := range arr4 {
		fmt.Printf("Index: %d, Value: %d\n", index, value)
	}
}

func slices_func() {
	// Khai báo — zero value của slice là nil (khác array!)
	var arr []string
	fmt.Println(arr == nil) // true

	// C1. make để tạo slice với độ dài và dung lượng
	slice1 := make([]string, 0, 5)
	fmt.Println(slice1) // []

	// C2. Khai báo + gán giá trị ngay
	slice2 := []string{"Go", "Python", "Java"}
	fmt.Println(slice2)

	// C3. cắt từ array
	arr2 := [5]int{1, 2, 3, 4, 5}
	slice3 := arr2[1:4] // slice từ index 1 đến 3 (4 không bao gồm)
	fmt.Println(slice3) // [2 3 4]

	// Thêm phần tử vào slice
	slice4 := []int{10, 20}
	slice4 = append(slice4, 30, 40)
	fmt.Println(slice4) // [10 20 30 40]

	// sao chép slice - copy
	src := []int{1, 2, 3, 4}
	dst := make([]int, len(src))
	copy(dst, src)
	fmt.Println(src)
	fmt.Println(dst)

}

func map_func() {
	// 1. Khai báo và khởi tạo
	m1 := make(map[string]int)
	fmt.Println(m1)

	m2 := map[string]int{
		"apple":  5,
		"banana": 3,
	}
	fmt.Println(m2)

	// 2. Các thao tác cơ bản
	m3 := map[string]int{"a": 1, "b": 2}
	m3["c"] = 3
	m3["d"] = 4

	fmt.Println(m3["c"])

	// 3. Check key có tồn tại không
	val, ok := m3["c"]
	if ok {
		fmt.Println("Có key c, giá trị", val)
	} else {
		fmt.Println("Không có key c")
	}

	// 4. Xóa
	delete(m3, "b")
	fmt.Println(m3)

	// 5. Duyệt
	for key, value := range m3 {
		fmt.Println(key, value)
	}
}

func main() {
	// array_func()
	// slices_func()
	map_func()
}
