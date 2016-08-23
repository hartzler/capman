package cmd

import (
	"fmt"
	"time"

	"github.com/hartzler/capman/state"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(&heartbeatCmd)
	heartbeatCmd.Flags().IntVar(&hearbeatConfig.interval, "interval", 10, "The number of seconds to sleep between heartbeats")
	heartbeatCmd.Flags().StringVar(&hearbeatConfig.livelinessCheckURL, "liveliness-check-url", "", "Liveliness check url")
	heartbeatCmd.Flags().StringVar(&hearbeatConfig.livelinessCheckTimeout, "liveliness-check-timeout", "", "Liveliness check url")
	heartbeatCmd.Flags().StringVar(&hearbeatConfig.bootstrap, "bootstrap", "", "command to run before quorum known")
	heartbeatCmd.Flags().IntVar(&hearbeatConfig.quorum, "quorum", 2, "The number of peers required for quorum")
	heartbeatCmd.Flags().StringVar(&hearbeatConfig.quorumInitial, "quorum-initial", "", "command to run once first quorum is achieved")
	heartbeatCmd.Flags().StringVar(&hearbeatConfig.quorumGained, "quorum-gained", "", "command to run when quorum achieved after the first time")
	heartbeatCmd.Flags().StringVar(&hearbeatConfig.quorumLost, "quorum-lost", "", "command to run when quorum is lost")
	heartbeatCmd.Flags().StringVar(&hearbeatConfig.peerJoin, "peer-join", "", "command to run when peer joins")
	heartbeatCmd.Flags().StringVar(&hearbeatConfig.peerLeave, "peer-leave", "", "command to run when peer leaves")
}

// --liveliness-check-url=http://localhost:4001
// --liveliness-check-timeout=10s
// --bootstrap="/opt/etcd/bootstrap.sh"
// --quorum-initial="/opt/etcd/quorum-initial.sh"
// --quorum-gained="/opt/etcd/quorum-gained.sh"
// --quorum-lost="/opt/etcd/quorum-lost.sh"
// --peer-join="/opt/etcd/peer-join.sh"
// --peer-leave="/opt/etcd/peer-left.sh"
// --quorum=3
type hbConfig struct {
	livelinessCheckURL     string
	livelinessCheckTimeout string
	bootstrap              string
	quorumInitial          string
	quorumGained           string
	quorumLost             string
	peerJoin               string
	peerLeave              string
	interval               int
	healthyPeerDuration    time.Duration
	quorum                 int
}

var hearbeatConfig hbConfig

type quorumState int

const (
	// QuorumUnknown for unknown state
	QuorumUnknown quorumState = iota
	// QuorumNever for when quorum has never been achieved
	QuorumNever
	// QuorumGained for any time after the first time quorum is achieved
	QuorumGained
	// QuorumLost for when quorum is lost
	QuorumLost
)

var lastKnownQuorumState quorumState

var heartbeatCmd = cobra.Command{
	Use:   "heartbeat",
	Short: "Post self into peer list",
	Run: func(cmd *cobra.Command, args []string) {
		state := stateFromContext(cmd)

		// see if quorum has been initialized
		lastKnownQuorumState = QuorumUnknown
		init, err := state.IsInitialized()
		if err != nil {
			panic(err)
		}

		eventLoop(state, init)
	},
}

func eventLoop(state state.ExternalState, init *state.Initialized) {
	if init == nil {
		lastKnownQuorumState = QuorumNever
	}

	// event loop
	for {
		// sleep
		time.Sleep(time.Duration(hearbeatConfig.interval) * time.Second)

		// report in
		if err := state.Heartbeat(); err != nil {
			fmt.Println("Heartbeat ERROR:", err)
			continue
		}
		fmt.Println("Updated heartbeat.")

		// get peer list
		peers, err := state.Peers()
		if err != nil {
			fmt.Println("Peers ERROR:", err)
			continue
		}

		// calculate quorum states
		healthy := 0
		for _, peer := range peers {
			if peer.IsHealthy(hearbeatConfig.healthyPeerDuration) {
				healthy++
			}
		}

		// event quorum initial state
		if lastKnownQuorumState == QuorumNever && healthy >= hearbeatConfig.quorum {
			// fire initialized command (at least once)
			if err := exec(hearbeatConfig.quorumInitial); err != nil {
				fmt.Println("quorumInitial exec ERROR:", err)
				continue
			}

			// set initialized state
			_, err = state.SetInitialized()
			if err != nil {
				fmt.Println("SetInitialized ERROR:", err)
				continue
			}

			lastKnownQuorumState = QuorumGained
		}

		// event quorum gained state
		if lastKnownQuorumState == QuorumLost && healthy >= hearbeatConfig.quorum {
			// fire quroum gained command
			if err := exec(hearbeatConfig.quorumGained); err != nil {
				fmt.Println("quorumGained exec ERROR:", err)
				continue
			}

			lastKnownQuorumState = QuorumGained
		}

		// event quorum lost state
		if lastKnownQuorumState == QuorumGained && healthy < hearbeatConfig.quorum {
			// fire quroum gained command
			if err := exec(hearbeatConfig.quorumLost); err != nil {
				fmt.Println("quorumLost exec ERROR:", err)
				continue
			}

			lastKnownQuorumState = QuorumLost
		}
	}
}

func exec(cmd string) error {
	return nil
}
