// remote state store for peer information and coordination
package state

type Peer struct {
  Host string
  Ip string
}

type Config struct {
  Me Peer
  Prefix string // the prefix for the remote state
}

type ExternalState interface {
  Heartbeat() error
  Peers() ([]Peer, error)
}
