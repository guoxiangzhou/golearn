package main

import "fmt"

func main() {
	defer fmt.Println("world!")
	fmt.Print("hello ")
}
