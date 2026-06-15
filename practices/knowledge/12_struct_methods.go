package main

import "fmt"

// 1. Struct là một kiểu dữ liệu tổng hợp (composite type) cho phép bạn nhóm nhiều giá trị có liên quan lại với nhau dưới một tên duy nhất. Mỗi giá trị trong struct được gọi là trường (field) và có thể có kiểu dữ liệu khác nhau.
type Person struct {
	Name string
	Age  int
}

func struct_func() {
	// Khởi tạo có tên trường (khuyên dùng)
	p := Person{Name: "Karan", Age: 22}
	// Khởi tạo không tên trường (phải đúng thứ tự và đủ số lượng)
	p2 := Person{"Bruce", 22}
	// Anonymous struct (Dùng một lần, không cần đặt tên kiểu)
	a := struct{ Title string }{"Golang"}

	println(p.Name, p.Age)
	println(p2.Name, p2.Age)
	println(a.Title)
}

// 2.Embedding and Composition: Go hỗ trợ embedding, cho phép bạn nhúng một struct vào struct khác. Điều này giúp tạo ra các cấu trúc phức tạp hơn mà không cần kế thừa truyền thống.

type Employee struct {
	Person   // Nhúng struct Person vào Employee
	Position string
}

func embedding_func() {
	e := Employee{
		Person:   Person{Name: "Alice", Age: 30},
		Position: "HR",
	}

	println(e.Name, e.Age, e.Position) // Truy cập trực tiếp các trường của Person
}

// 3. Struct Methods: Bạn có thể định nghĩa các phương thức (methods) cho struct, cho phép bạn gắn hành vi vào dữ liệu của struct.
// format: func (receiver type) methodName(parameters) returnType { ... }

type Rect struct {
	Width, Height float64
}

// Value receiver: chỉ tính toán, không sửa đổi dữ liệu gốc
func (r Rect) Area(scale float64) float64 {
	r.Width *= scale  // no change
	r.Height *= scale // no change
	return 3.14 * r.Width * r.Height
}

// Pointer receiver: có thể sửa đổi dữ liệu gốc
func (r *Rect) Scale(scale float64) {
	fmt.Printf("Before width and height: %d %d\n", r.Width, r.Height)
	r.Width *= scale
	r.Height *= scale
	fmt.Printf("After width and height: %d %d\n", r.Width, r.Height)
}

func main() {
	// struct_func()
	// embedding_func()
	r := Rect{Width: 5, Height: 10}

	println(r.Area(0.2))
	r.Scale(0.2)
}
