 package cmd

import (
  "fmt"
  "github.com/spf13/cobra"
)

func init() {
  RootCmd.AddCommand(&heartbeatCmd)
}

// --liveliness-check-url=http://localhost:4001
// --liveliness-check-timeout=10s
// --bootstrap="/opt/etcd/bootstrap.sh"
// --quorum-initial="/opt/etcd/quorum-initial.sh"
// --quorum-gained="/opt/etcd/quorum-gained.sh"
// --quorum-lost="/opt/etcd/quorum-lost.sh"
// --peer-join="/opt/etcd/peer-join.sh"
// --peer-leave="/opt/etcd/peer-left.sh"
type hbConfig struct {
  livelinessCheckUrl string
  livelinessCheckTimeout string
  bootstrap string
  quorumInitial string
  quorumGained string
  quorumLost string
  peerJoin string
  peerLeave string
  interval time.Duration
  healthyPeerDuration time.Duration
  quorum int
}

var hearbeatConfig hbConfig

type quorumState int
const (
  QUORUM_UNKNOWN quorumState = iota
  QUORUM_INITIAL
  QUORUM_GAINED
  QUORUM_LOST
)

var lastKnownQuorumState quorumState

var heartbeatCmd = cobra.Command{
  Use: "heartbeat",
  Short: "Post self into peer list",
  Run: func(cmd *cobra.Command, args []string) {
    state := stateFromContext(cmd)

    // see if we quorum has been initialized
    init, err := state.IsInitialized()
    if err != nil {
      panic(err)
    }

    // event loop
    for {

      // get peer list
      peers, err := state.Peers()
      if err != nil {
        panic(err) // TODO: better
      }

      // calculate quorum states
      healthy := 0
      for _, peer := peers {
        if peer.IsHealthy(hearbeatConfig.healthyPeerDuration) {
          healthy += 1
        }
      }

      // initial
      if init == nil && healthy >= hearbeatConfig.quorum
        // fire initialized command (at least once)
        if err := exec(hearbeatConfig.quorumInitial); err != nil {
          panic(err) // TODO: better
        }

        // set initialized state
        init, err := state.SetInitialized()
        if err != nil {
          panic(err) // TODO: better
        }

        quarumEvented = true
      }

      // quorum gained state
      if init != nil && healthy >= hearbeatConfig.quorum
        // fire quroum gained command
        if err := exec(hearbeatConfig.quorumGained); err != nil {
          panic(err) // TODO: better
        }
      }

      // report in
      err := state.Heartbeat()
      if err != nil {
        panic(err) // TODO: better
      }
      fmt.Println("Updated heartbeat.")

      // sleep
      time.Sleep(hearbeatConfig.interval)
    }
  },
}

heartbeatCmd.Flags().StringVar(&hearbeatConfig.livelinessCheckUrl, "liveliness-check-url", "", "Liveliness check url")
heartbeatCmd.Flags().StringVar(&hearbeatConfig.livelinessCheckTimeout, "liveliness-check-timeout", "", "Liveliness check url")
heartbeatCmd.Flags().StringVar(&hearbeatConfig.bootstrap, "bootstrap", "", "command to run before quorum known")
heartbeatCmd.Flags().IntVar(&hearbeatConfig.quorum, "quorum", "", "The number of peers required for quotum")
heartbeatCmd.Flags().StringVar(&hearbeatConfig.quorumInitial, "quorum-initial", "", "command to run once first quorum is achieved")
heartbeatCmd.Flags().StringVar(&hearbeatConfig.quorumGained, "quorum-gained", "", "command to run when quorum achieved after the first time")
heartbeatCmd.Flags().StringVar(&hearbeatConfig.quorumLost, "quorum-lost", "", "command to run when quorum is lost")
heartbeatCmd.Flags().StringVar(&hearbeatConfig.peerJoin, "peer-join", "", "command to run when peer joins")
heartbeatCmd.Flags().StringVar(&hearbeatConfig.peerLeave, "peer-leave", "", "command to run when peer leaves")
