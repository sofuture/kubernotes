package main

import (
	"fmt"
	"os"
)

func main() {
	err := runCli()

	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
}
