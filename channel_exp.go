package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan string)
	defer close(ch)
	go func(ch <-chan string) {
		v := <-ch
		for v != "stop" {
			fmt.Println(v)
			v = <-ch
		}
		fmt.Println("exit receiver")
	}(ch)

	go func(ch chan<- string) {
		ch <- "hello"
		time.Sleep(1e9)
		ch <- "world"
		time.Sleep(1e9)
		ch <- "fuck"
		time.Sleep(1e9)
		ch <- "stop"
		fmt.Println("exit sender")
	}(ch)
	time.Sleep(5e9)
	fmt.Println("exit main")
}
