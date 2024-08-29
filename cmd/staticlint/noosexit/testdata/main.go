package main

import (
	"fmt"
	"os"
)

func notMainFunc() {
	os.Exit(2)
}

func main() {
	fmt.Println("Main package with main function")
	notMainFunc()
	os.Exit(0)
}
