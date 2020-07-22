package main

import (
	"fmt"
	"time"
)

func foo() {
	a := 10
	fmt.Printf("begin a = %d\n", a)
	go func() {
		fmt.Printf("begin goo a = %d\n", a)
		a = 20
		time.Sleep(1e9)
		fmt.Printf("end goo a = %d\n", a)
	}()
	time.Sleep(0.5e9)
	fmt.Printf("end a = %d\n", a)
}

func main() {
	foo()
	time.Sleep(2e9)
}
