package main

import "fmt"

func f1() {
	x := 0
	fmt.Printf("init x = %d\n", x)
	func() {
		fmt.Printf("begin lambda x = %d\n", x)
		x = 1
		fmt.Printf("end lambda x = %d\n", x)
	}()
	fmt.Printf("after lambda x = %d\n", x)
}

func f2() {
	x := 0
	fmt.Printf("init x = %d\n", x)
	func(x interface{}) {
		fmt.Printf("begin lambda x = %d\n", x)
		x = 1
		fmt.Printf("end lambda x = %d\n", x)
	}(x)
	fmt.Printf("after lambda x = %d\n", x)
}

//func f3() {
//	x := 0
//	fmt.Printf("init x = %d\n", x)
//	func(x interface{}) {
//		fmt.Printf("begin lambda x = %d\n", *x)
//		x = 1
//		fmt.Printf("end lambda x = %d\n", *x)
//	}(&x)
//	fmt.Printf("after lambda x = %d\n", x)
//}

func main() {
	f1()
	f2()
	//f3()
}
