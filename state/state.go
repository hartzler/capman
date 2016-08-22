// remote state store for peer information and coordination
package state

import (
  "time"
)

type Peer struct {
  Host string
  Ip string
  LastSeen time.Time
}

func (self Peer) IsHealthy(duration time.Duration) bool {
  healthyT := time.Now().Add(-duration)
  return self.LastSeen.After(healthyT)
}

type Initialized struct {
  First time.Time
}

type Config struct {
  Me Peer
  Prefix string // the prefix for the remote state
}

type ExternalState interface {
  Heartbeat() error
  Peers() ([]Peer, error)
}
