package cmd

import (
  "fmt"
  "github.com/spf13/cobra"
)

func init() {
  RootCmd.AddCommand(&heartbeatCmd)
}

var heartbeatCmd = cobra.Command{
  Use: "heartbeat",
  Short: "Post self into peer list",
  Run: func(cmd *cobra.Command, args []string) {
    state := stateFromContext(cmd)
    err := state.Heartbeat()
    if err != nil {
      panic(err)
    }
    fmt.Println("Updated heartbeat.")
  },
}
