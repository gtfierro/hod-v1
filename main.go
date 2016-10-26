package main

import (
	"./goraptor"
	"fmt"
	"os"
)

func main() {
	file := os.Args[1]
	p := turtle.NewParser(file)
	for _, t := range p.Triples {
		fmt.Println(t)
	}
	for pfx, ns := range p.Namespaces {
		fmt.Println(pfx, "=>", ns)
	}
}
