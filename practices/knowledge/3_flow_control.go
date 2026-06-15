package main

import "fmt"

func main() {
	// 1. If/Else
	var x int = 5

	if x > 10 {
		fmt.Println("x lớn hơn 10")
	} else {
		fmt.Println("x nhỏ hơn 10")
	}

	// 2. Compact if
	if x := 10; x > 5 {
		fmt.Println("x lớn hơn 5")
	} else {
		fmt.Println("x nhỏ hơn 5")
	}

	// 3. Switch case
	day := "monday"

	switch day {
	case "monday":
		fmt.Println("It's Monday")
	case "tuesday":
		fmt.Println("It's Tuesday")
	default:
		fmt.Println("It's another day")
	}

}
