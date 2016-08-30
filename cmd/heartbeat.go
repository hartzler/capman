package cmd

import (
	"fmt"
	"net/http"
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
	heartbeatCmd.Flags().IntVar(&hearbeatConfig.healthyPeerDuration, "healthy-peer-duration", 60, "The number of seconds a peer is healthy for")
	heartbeatCmd.Flags().StringVar(&hearbeatConfig.quorumGained, "quorum-gained", "", "command to run when quorum achieved after the first time")
	heartbeatCmd.Flags().StringVar(&hearbeatConfig.quorumLost, "quorum-lost", "", "command to run when quorum is lost")
	heartbeatCmd.Flags().StringVar(&hearbeatConfig.peerJoin, "peer-join", "", "command to run when peer joins")
	heartbeatCmd.Flags().StringVar(&hearbeatConfig.peerLeave, "peer-leave", "", "command to run when peer leaves")
}

// --liveliness-check-url=http://localhost:4001
// --liveliness-check-timeout=10s
// --bootstrap="/opt/etcd/bootstrap.sh"
// --quorum-gained="/opt/etcd/quorum-gained.sh"
// --quorum-lost="/opt/etcd/quorum-lost.sh"
// --peer-join="/opt/etcd/peer-join.sh"
// --peer-leave="/opt/etcd/peer-left.sh"
// --quorum=2
type hbConfig struct {
	livelinessCheckURL     string
	livelinessCheckTimeout string
	bootstrap              string
	quorumGained           string
	quorumLost             string
	peerJoin               string
	peerLeave              string
	interval               int
	healthyPeerDuration    int
	quorum                 int
}

var hearbeatConfig hbConfig

type clusterState int

const (
	// ClusterUnknown for unknown state
	ClusterUnknown clusterState = iota
	// ClusterBootstrap for when quorum has never been achieved
	ClusterBootstrap
	// ClusterQuorum for any time after the first time quorum is achieved
	ClusterQuorum
	// ClusterLost for when quorum is lost
	ClusterLost
)

var currentState = ClusterUnknown

func healthySeconds() time.Duration {
	return time.Duration(hearbeatConfig.healthyPeerDuration) * time.Second
}

var heartbeatCmd = cobra.Command{
	Use:   "heartbeat",
	Short: "Post self into peer list",
	Run: func(cmd *cobra.Command, args []string) {
		state := stateFromContext(cmd)

		// determine initial state
		currentState = ClusterUnknown
		boot, err := state.IsBootstrap()
		if err != nil {
			panic(err)
		}
		if boot == nil {
			currentState = ClusterBootstrap
		} else {
			// get peer list
			peers, err := state.Peers()
			if err != nil {
				panic(err)
			}
			// see if we have or lost quorum
			if len(peers.Healthy(healthySeconds())) >= hearbeatConfig.quorum {
				currentState = ClusterQuorum
			} else {
				currentState = ClusterLost
			}
		}

		fmt.Println("currentState:", currentState)

		// now that we have our initial state, do the event loop
		eventLoop(state)
	},
}

func eventLoop(state state.ExternalState) {
	for {
		// sleep (so we can just continue on error and not hammer)
		time.Sleep(time.Duration(hearbeatConfig.interval) * time.Second)

		// skip health checks till we are boostrapped
		if currentState != ClusterBootstrap && hearbeatConfig.livelinessCheckURL != "" {
			resp, err := http.Get(hearbeatConfig.livelinessCheckURL)
			if err != nil {
				fmt.Println("Health check fail!", err)
				continue
			}
			if resp.StatusCode != 200 {
				fmt.Println("Health check fail!", err)
				continue
			}
		}
		// report in
		if err := state.Heartbeat(); err != nil {
			fmt.Println("Heartbeat ERROR:", err)
			continue
		}

		// get peer list
		peers, err := state.Peers()
		if err != nil {
			fmt.Println("Peers ERROR:", err)
			continue
		}

		// calculate quorum states
		healthy := 0
		for _, peer := range peers {
			health := "unhealthy"
			if peer.IsHealthy(healthySeconds()) {
				health = "healthy"
				healthy++
			}
			fmt.Println(peer.Host, peer.IP, health)
		}

		// event bootstrap
		if currentState == ClusterBootstrap && healthy >= hearbeatConfig.quorum {
			fmt.Println("EVENT: bootstrap!")
			// fire bootstrap command (at least once)
			if err := exec(hearbeatConfig.bootstrap); err != nil {
				fmt.Println("bootstrap exec ERROR:", err)
				continue
			}

			// set cluster state to bootstrap
			_, err = state.SetBootstrap()
			if err != nil {
				fmt.Println("SetInitialized ERROR:", err)
				continue
			}
			currentState = ClusterQuorum

		}

		// event gained
		if currentState == ClusterLost && healthy >= hearbeatConfig.quorum {
			// event quorum gained state
			fmt.Println("EVENT: quorumGained")
			// fire quroum gained command
			if err := exec(hearbeatConfig.quorumGained); err != nil {
				fmt.Println("quorumGained exec ERROR:", err)
				continue
			}
			currentState = ClusterQuorum
		}

		// event quorum lost state
		if currentState == ClusterQuorum && healthy < hearbeatConfig.quorum {
			fmt.Println("EVENT: quorumLost")
			// fire quroum gained command
			if err := exec(hearbeatConfig.quorumLost); err != nil {
				fmt.Println("quorumLost exec ERROR:", err)
				continue
			}
			currentState = ClusterLost
		}
	}
}

func exec(cmd string) error {
	fmt.Println("DEBUG: todo exec:", cmd)
	return nil
}
