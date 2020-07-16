package main

import (
	"flag"
	"fmt"
)

func main() {
	strflag := flag.String("strflag", "default-string-val", "string flag")
	intflag := flag.Int("intflag", 12, "int flag")
	flag.Parse()
	fmt.Printf("str val = %s, int val = %d\n", *strflag, *intflag)
}
