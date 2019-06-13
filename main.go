package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Key service started")

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("No keys")
		os.Exit(1)
	}

	fmt.Print("Received")
	fmt.Print(args)
}
