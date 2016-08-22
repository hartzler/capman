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

func (self *consul) prefix(s string) string {
  return fmt.Sprintf("%s/%s", self.config.Prefix, s)
}

func (self *consul) Heartbeat() error {
  fmt.Println("Posting my info to peers: ", self.config.Me)
  kv := self.client.KV()
  key := self.prefix(fmt.Sprintf("peers/%s", self.config.Me.Host))
  self.config.Me.LastSeen = time.Now()
  bytes, err := json.Marshal(self.config.Me)
  if err != nil {
    return err
  }
  _, err = kv.Put(&api.KVPair{Key: key, Value: bytes}, nil)
  return err
}

func (self *consul) IsInitialized() (Initialized, error) {
  fmt.Println("Checking if initial quorum was ever reached...")
  kv := self.client.KV()
  pair, _, err := kv.Get(self.prefix("initialized"), nil)
  if err != nil {
    return nil, err
  }
  var init Initilaized
  err := json.Unmarshal(pair.Value, &init)
  return init, err
}

func (self *consul) SetInitialized() (Initialized, error) {
  fmt.Println("Setting first quorum achieved...")
  kv := self.client.KV()
  init := Initialized{time.Now()}
  bytes, err := json.Marshal(init)
  if err != nil {
    return nil, err
  }
  _, err = kv.Put(&api.KVPair{Key: self.prefix("initialized"), Value: bytes}, nil)
  return init, err
}

func (self *consul) Peers() ([]Peer, error) {
  fmt.Println("Retrieving list of peers...")
  kv := self.client.KV()
  pairs, _, err := kv.List(self.prefix("peers"), nil)
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
