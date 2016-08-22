// Package state stores and retrieves the peer information for coordination.
// Different backend stores should be supported.  Currently only Consul is.
package state

import (
	"time"
)

// Peer represents the state of a peer
type Peer struct {
	Host     string
	IP       string
	LastSeen time.Time
}

// IsHealthy determines if this peer is in a healthy state.
func (p Peer) IsHealthy(duration time.Duration) bool {
	healthyT := time.Now().Add(-duration)
	return p.LastSeen.After(healthyT)
}

// Initialized stores information about when the cluster was
type Initialized struct {
	First time.Time
}

// Config is the high level settings for running capman
type Config struct {
	Me     Peer
	Prefix string // the prefix for the remote state
}

// ExternalState is the interface for dealing with the external state store.
type ExternalState interface {
	IsInitialized() (Initialized, error)
	SetInitialized() (Initialized, error)
	Heartbeat() error
	Peers() ([]Peer, error)
}
