package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan string)
	go func(ch <-chan string) {
		for v := range ch {
			fmt.Println(v)
		}

		//v,ok := <-ch
		//for ok {
		//	fmt.Println(v)
		//	v, ok = <-ch
		//}
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
		close(ch)
		fmt.Println("exit sender")
	}(ch)
	time.Sleep(5e9)
	fmt.Println("exit main")
}
