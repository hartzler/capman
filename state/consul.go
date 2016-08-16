// consul backend driver for storing external state of peers
package state

import (
  "encoding/json"
  "fmt"
  "time"
  "github.com/hashicorp/consul/api"
)

type consul struct {
  config Config
  consul ConsulConfig
  client *api.Client
}

type ConsulConfig struct {
	address    string
	token      string
	quiet      bool
	waitIndex  uint64
	consistent bool
	stale      bool
}

func NewConsul(config Config, consulConfig ConsulConfig) ExternalState {
  // Get a new client
  client, err := api.NewClient(api.DefaultConfig())
  if err != nil {
      panic(err)
  }
  return &consul{
    config: config,
    consul: consulConfig,
    client: client,
  }
}

func (self *consul) Heartbeat() error {
  fmt.Println("Posting my info to peers: ", self.config.Me)
  kv := self.client.KV()
  key := fmt.Sprintf("%s/peers/%s", self.config.Prefix, self.config.Me.Host)
  self.config.Me.LastSeen = time.Now()
  bytes, err := json.Marshal(self.config.Me)
  if err != nil {
    return err
  }
  _, err = kv.Put(&api.KVPair{Key: key, Value: bytes}, nil)
  return err
}

func (self *consul) Peers() ([]Peer, error) {
  fmt.Println("Retrieving list of peers...")
  kv := self.client.KV()
  pairs, _, err := kv.List(fmt.Sprintf("%s/peers", self.config.Prefix), nil)
  if err != nil {
    return nil, err
  }
  peers := make([]Peer, len(pairs))
  for i, pair := range pairs {
    var peer Peer
    err := json.Unmarshal(pair.Value, &peer)
    if err != nil {
      fmt.Println("Error decoding peer record:s", pair.Key, "skipping...", err)
      continue
    }
    peers[i] = peer
  }
  return peers, nil
}
