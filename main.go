package main

import (
	"fmt"
	"os"

	"github.com/hartzler/capman/cmd"
)

const Name = "capman"
const Version = "0.1.0"

func main() {
	root := cmd.Init(Name, Version)
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintln(root.Err, err)
		os.Exit(1)
	}
}
