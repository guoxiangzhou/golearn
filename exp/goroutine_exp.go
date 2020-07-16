package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Begin main")
	go shortwait()
	go longWait()
	time.Sleep(10 * 1e9)
	fmt.Println("End main")
}

func longWait() {
	fmt.Println("Begin long wait")
	time.Sleep(5 * 1e9)
	fmt.Println("End long wait")
}

func shortwait() {
	fmt.Println("Being short wait")
	time.Sleep(1 * 1e9)
	fmt.Println("End short wait")
}
