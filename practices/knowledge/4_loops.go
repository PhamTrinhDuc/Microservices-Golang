package main

import "fmt"

func main() {
	fmt.Println("1. For loop")
	for i := 0; i <= 10; i++ {
		fmt.Println(i)
	}

	fmt.Println("2. For loop as while loop")
	for i := 0; i <= 10; {
		fmt.Println(i)
		i++
	}

	fmt.Println("3. For loop as infinite loop")
	for {
		fmt.Println("Infinite loop")
		break
	}

	fmt.Println("4. For loop with break")
	for i := 0; i <= 10; i++ {
		if i == 5 {
			break
		}
		fmt.Println(i)
	}

	fmt.Println("5. For loop with continue")
	for i := 0; i <= 10; i++ {
		if i == 5 {
			continue
		}
		fmt.Println(i)
	}

	fmt.Println("6. For loop with goto")
	for i := 0; i <= 10; i++ {
		if i == 5 {
			goto end
		}
		fmt.Println(i)
	}

end:
	fmt.Println("End")

	fmt.Println("7. For loop with nested loop")
	for i := 0; i <= 5; i++ {
		for j := 0; j <= 5; j++ {
			fmt.Println(i, j)
		}
	}

	fmt.Println("8. For loop with range")
	for i := range 5 {
		fmt.Println(i)
	}
}
