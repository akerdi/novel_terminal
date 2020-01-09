package main

import (
	"fmt"
	"os"

	"novel/cmd"
)

func main() {
	fmt.Println("Welcome to novel world!")
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
