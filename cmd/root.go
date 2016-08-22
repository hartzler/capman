package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/hartzler/capman/state"
	"github.com/spf13/cobra"
)

var version string
var name string

var RootCmd = &cobra.Command{
	Use:   "capman",
	Short: "Capman maintains an evented set of peers in consul.",
	Long:  `See http://github.com/hartzler/capman`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

var config = state.Config{
	Me: state.Peer{
		Host: "localhost",
		Ip:   "127.0.0.1",
	},
	Prefix: "k8s/master/runtime/etcd",
}

type Cmd struct {
	root *cobra.Command
	Err  io.Writer
	Out  io.Writer
}

func Init(name, version string) *Cmd {
	// add version command
	RootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number of capman",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("capman version", version)
		},
	})

	RootCmd.PersistentFlags().StringVar(&config.Prefix, "prefix", "p", "prefix for peer state")
	RootCmd.PersistentFlags().StringVar(&config.Me.Host, "host", "", "hostname for heartbeat")
	RootCmd.PersistentFlags().StringVar(&config.Me.Ip, "ip", "", "ip for heartbeat")

	return &Cmd{
		root: RootCmd,
		Err:  os.Stderr,
		Out:  os.Stdout,
	}
}

// configure the external state or exit out if invalid options specified
func stateFromContext(cmd *cobra.Command) state.ExternalState {
	consul := state.ConsulConfig{}
	return state.NewConsul(config, consul)
}
