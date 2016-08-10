package cmd

import (
  "fmt"
  "github.com/spf13/cobra"
)

func init() {
  RootCmd.AddCommand(&peersCmd)
}

var peersCmd = cobra.Command{
  Use: "peers",
  Short: "Get current list of peers",
  Run: func(cmd *cobra.Command, args []string) {
    state := stateFromContext(cmd)
    peers, err := state.Peers()
    if err != nil {
      panic(err)
    }
    fmt.Println("Peers: ", peers)
  },
}
