package main

import (
	"fmt"
	"time"
)

func main() {
	c := make(chan int, 100)
	go func(c <-chan int) {
		time.Sleep(5e9)
		fmt.Println("start receive from channel")
		for v := range c {
			fmt.Println(v)
		}
		fmt.Println("receive from channel finished")
	}(c)

	c <- 1
	c <- 2
	c <- 3
	fmt.Println("send three value into channel")
	time.Sleep(10e9)
	close(c)
	time.Sleep(1e9)
	fmt.Println("exit main")
}
