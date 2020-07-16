package main

import "fmt"

func main() {
	c1 := make(chan string)
	defer close(c1)

	c2 := make(chan string)
	defer close(c2)
	go func(c chan<- string) {
		c <- "test 1"
	}(c1)
	go func(c chan<- string) {
		c <- "test 2"
	}(c2)
	for i := 0; i < 2; i++ {
		select {
		case v := <-c1:
			fmt.Printf("idx = %d, value = %s\n", i, v)
		case v := <-c2:
			fmt.Printf("idx = %d, value = %s\n", i, v)
		}
	}
}
