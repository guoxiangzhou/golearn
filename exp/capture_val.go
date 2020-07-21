package main

import "fmt"

func main() {
	x := 0
	fmt.Printf("init x = %d\n", x)
	func() {
		fmt.Printf("begin lambda x = %d\n", x)
		x = 1
		fmt.Printf("end lambda x = %d\n", x)
	}()
	fmt.Printf("after lambda x = %d\n", x)
}
