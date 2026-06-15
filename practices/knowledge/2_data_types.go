package main

import "fmt"

func main() {

	// 1. Numerical data types
	var i = 10
	fmt.Printf("Type of x: %T\n", i)
	fmt.Printf("i: %d\n", i)

	// 2. String data types
	// one line
	var foo = "Hello world"
	fmt.Printf(foo + "\n")
	// multiple line
	var bio string = `I am statically typed.
										I was designed at Google.`
	fmt.Printf(bio + "\n")

	// 3. boolean
	var is_value = true
	fmt.Printf("is_value: %t\n", is_value)

	// 4. Operators (!, &&, ==, !=)
	var num1 = 10
	var num2 = 20
	fmt.Printf("Equal to: %t\n", num1 == num2)
	fmt.Printf("Not equal to: %t\n", num1 != num2)
	fmt.Printf("Greater than: %t\n", num1 > num2)
	fmt.Printf("Less than: %t\n", num1 < num2)
	fmt.Printf("Greater than or equal to: %t\n", num1 >= num2)
	fmt.Printf("Less than or equal to: %t\n", num1 <= num2)

	// 5. Type Conversion
	var num_int int = 64
	var num_float float32 = float32(num_int)
	fmt.Printf("num_float: %f\n", num_float)
	fmt.Printf("num_int: %d\n", num_int)

	// 6. Alias types
	type Celsius float64
	var temperature Celsius = 25.0
	fmt.Printf("temperature: %f\n", temperature)
	fmt.Printf("Type of temperature: %T\n", temperature)

	// sự khác biệt giữa Type Alias (Tên giả) vs. Defined Type (Kiểu định nghĩa mới).
	// Type Alias: Chỉ là một cái tên khác cho kiểu dữ liệu đã có.
	// Defined Type: Tạo ra một kiểu dữ liệu hoàn toàn mới, không tương thích với kiểu gốc.
	type MyAlias = string
	type MyDefined string
	var alias MyAlias = "Hello"
	var defined MyDefined = "Word"

	var copy1 string = alias // work: MyAlias chỉ là một cái tên khác cho string
	fmt.Printf("copy1: %s\n", copy1)
	var copy2 string = string(defined) // error: không thể gán MyDefined cho string vì MyDefined là một kiểu dữ liệu hoàn toàn mới, cần ép sang string
	fmt.Printf("copy2: %s\n", copy2)
}
