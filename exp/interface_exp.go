package main

import "fmt"

type Base interface {
	Print()
}

type Derived1 struct {
}

func (d *Derived1) Print() {
	fmt.Println("from derived1")
}

type Derived2 struct {
}

func (d *Derived2) Print() {
	fmt.Println("from derived2")
}

func PrintTest(p Base) {
	p.Print()
}

func main() {
	d1 := &Derived1{}
	d2 := &Derived2{}
	PrintTest(d1)
	PrintTest(d2)
}
