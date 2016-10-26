package main

import (
	"./goraptor"
	"os"
)

func main() {
	file := os.Args[1]
	turtle.ParseFile(file)
}
