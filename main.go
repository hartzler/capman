package main

import (
	"fmt"
	//"io/ioutil"
	//"log"
	"os"

	"github.com/hartzler/capman/cmd"
)

const Name = "capman"
const Version = "0.1.0"

func main() {
	//log.SetOutput(ioutil.Discard)

	root := cmd.Init(Name, Version)
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintln(root.Err, err)
		os.Exit(1)
	}

	os.Exit(0)
}
