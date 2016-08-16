package cmd

import (
  "fmt"
  "time"
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
    healthyT := time.Now().Add(-time.Second*60)
    fmt.Println("Peers: ")
    for _, peer := range peers {
      health := "unhealthy"
      if peer.LastSeen.After(healthyT) {
        health = "healthy"
      }
      fmt.Println(peer.Host, peer.Ip, health)
    }
  },
}
