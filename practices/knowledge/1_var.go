package main

import "fmt"

func main() {

	// 1. Declare and initialize variables
	var a, b = 10, 20
	fmt.Printf("a: %d, b: %d\n", a, b)

	var (
		foo string = "Hello"
		bar string = "World"
	)

	fmt.Printf("foo: %s, bar %s\n", foo, bar)

	// 2. Short variable declaration
	x, y := 30, 40
	fmt.Printf("x %d, y %d\n", x, y)

	// 3. Constants
	const pi = 3.14
	fmt.Printf("pi: %f\n", pi)

	// Assigning new values to variables
	// x = 50       // => work
	// pi = 3.14159 // => compile error: cannot assign to pi (constant)

}

// run: go run 1_var.go
