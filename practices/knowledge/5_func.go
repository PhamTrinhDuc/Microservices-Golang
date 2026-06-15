package main

import "fmt"

// 1. Return value
func sum_caculate(a int, b int) int {
	var c int = a + b
	return c
}

// 2. Return multiple values
func multiple_devide_caculate(a int, b int) (int, int) {
	var c int = a * b
	var d int = a / b
	return c, d
}

// 3. Return named values
func return_named_values(a int, b int) (sum int, multiple int, devide int) {
	sum = a + b
	multiple = a * b
	devide = a / b
	return
}

// 4. Functions as values
func return_func() {
	fn := func() {
		fmt.Println("inside fn")
	}

	fn()
}

// 5. Closures
func myFunc() func(int) int {
	sum := 0

	return func(x int) int {
		sum += x
		return sum
	}
}

// 6. Variadic Functions
func add(values ...int) int {
	sum := 0
	for _, v := range values {
		sum += v
	}
	return sum
}

// 7. Init
func init() {
	fmt.Println("Init first function")
}

func init() {
	fmt.Println("Init second function")
}

// 8. Defer
func defer_func() {
	// Các steps thực hiện
	// 1. Thực thi trong hàm chính: print(Giá trị của a...)
	// 2. Nhận kết quả của hàm con là sub_defer và thực thi kết quả + defer trong hàm này
	// 3. Thực thi phần defer trong hàm chính

	sub_defer := func() {
		defer fmt.Println("1. Đóng cửa")
		defer fmt.Println("2. Tắt đèn")
		defer fmt.Println("3. Cất chìa khóa")

		fmt.Println("--- Đang làm việc trong phòng ---")
	}

	a := 10
	defer fmt.Println("Giá trị của a là:", a) // Go "chụp ảnh" a = 10 ngay lúc này

	a = 20
	fmt.Println("Giá trị hiện tại của a là:", a)

	sub_defer()
}

func main() {
	// var a int = 20
	// var b int = 10

	// 1. =========== RETURN VALUE
	// sum := sum_caculate(a, b)
	// fmt.Println("sum = ", sum)

	// 2. =========== RETURN MULTIPLE VALUES
	// var multiple, devide = multiple_devide_caculate(a, b)
	// fmt.Println("multiple = ", multiple)
	// fmt.Println("devide = ", devide)

	// 3. =========== RETURN NAMED VALUES
	// var sum, multiple, devide = return_named_values(a, b)
	// fmt.Println("sum = ", sum)
	// fmt.Println("multiple = ", multiple)
	// fmt.Println("devide = ", devide)

	// 4. =========== FUNCTIONS AS VALUES
	// return_func()

	// 5. =========== CLOSURES
	// var add = myFunc()
	// add(5)                          // => sum = 5
	// add(5)                          // => sum = 10
	// fmt.Printf("Sum %d\n", add(10)) // sum = 20

	// 6. ============== VARADIC FUNCTIONS
	// fmt.Printf("Sum %d\n", add(1, 2, 3, 4, 5))

	// 7. ============== DEFER
	defer_func()
}
