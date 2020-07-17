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

type derived3 struct {
}

func (d *derived3) NewPrint() {
	fmt.Println("from derived3")
}

type derived4 struct {
	derived2
}

type derived5 struct {
	derived2
}

func (d *derived5) Print() {
	fmt.Println("from derived5")
}

func PrintTest(p base) {
	p.Print()
}

func main() {
	d1 := &derived1{}
	d2 := &derived2{}
	PrintTest(d1)
	PrintTest(d2)
	//PrintTest(&derived3{})
	PrintTest(&derived4{})
	PrintTest(&derived5{})

	var p base
	p = &derived1{}
	p.Print()
	p = &derived2{}
	p.Print()
	//p = &derived3{}
	p = &derived4{}
	p.Print()
	p = &derived5{}
	p.Print()
}
