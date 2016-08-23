package main

import (
	"fmt"
	"os"

	"github.com/hartzler/capman/cmd"
)

// Name is the name of the executable
const Name = "capman"

// Version is the version of the executable
const Version = "0.1.0"

func main() {
	root := cmd.Init(Name, Version)
	if err := root.Execute(); err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}
}
