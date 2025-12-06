package main

import (
	"fmt"
	"os"
)

func main() {
	a := 2 + 3
	fmt.Println(a)
	os.Exit(0) // want "direct call of Exit"
}
