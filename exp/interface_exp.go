package main

import "fmt"

type base interface {
	Print()
}

type derived1 struct {
}

func (d *derived1) Print() {
	fmt.Println("from derived1")
}

type derived2 struct {
}

func (d *derived2) Print() {
	fmt.Println("from derived2")
}

func PrintTest(p base) {
	p.Print()
}

func main() {
	d1 := &derived1{}
	d2 := &derived2{}
	PrintTest(d1)
	PrintTest(d2)

	var p base
	p = &derived1{}
	p.Print()
	p = &derived2{}
	p.Print()
}
