package main

import (
	"./turtle"
	"os"
)

func main() {
	file := os.Args[1]
	f, e := os.Open(file)
	if e != nil {
		panic(e)
	}
	turtle.Parse(f)
}
