package main

import "fmt"

type baseTest struct {
	value int
}

func (b *baseTest) Set(v int) {
	b.value = v
}

func (b baseTest) Set2(v int) {
	b.value = v
}

func main() {
	val := baseTest{value: 12}
	fmt.Printf("before set, value = %d\n", val.value)
	val.Set(27)
	fmt.Printf("after set, value = %d\n", val.value)
	val.Set2(30)
	fmt.Printf("after set2, value = %d\n", val.value)
}
