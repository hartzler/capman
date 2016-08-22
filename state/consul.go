package state

// consul backend driver for storing external state of peers

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

// ConsulConfig is how to connect to the consul agent
type ConsulConfig struct {
	address    string
	token      string
	quiet      bool
	waitIndex  uint64
	consistent bool
	stale      bool
}

// NewConsul creates a new consul impl of ExternalState
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

func (c *consul) prefix(s string) string {
	return fmt.Sprintf("%s/%s", c.config.Prefix, s)
}

func (c *consul) Heartbeat() error {
	fmt.Println("Posting my info to peers: ", c.config.Me)
	kv := c.client.KV()
	key := c.prefix(fmt.Sprintf("peers/%s", c.config.Me.Host))
	c.config.Me.LastSeen = time.Now()
	bytes, err := json.Marshal(c.config.Me)
	if err != nil {
		return err
	}
	_, err = kv.Put(&api.KVPair{Key: key, Value: bytes}, nil)
	return err
}

func (c *consul) IsInitialized() (Initialized, error) {
	fmt.Println("Checking if initial quorum was ever reached...")
	var i Initialized
	kv := c.client.KV()
	pair, _, err := kv.Get(c.prefix("initialized"), nil)
	if err != nil {
		return i, err
	}
	err = json.Unmarshal(pair.Value, &i)
	return i, err
}

func (c *consul) SetInitialized() (Initialized, error) {
	fmt.Println("Setting first quorum achieved...")
	kv := c.client.KV()
	i := Initialized{time.Now()}
	bytes, err := json.Marshal(i)
	if err != nil {
		return i, err
	}
	_, err = kv.Put(&api.KVPair{Key: c.prefix("initialized"), Value: bytes}, nil)
	return i, err
}

func (c *consul) Peers() ([]Peer, error) {
	fmt.Println("Retrieving list of peers...")
	kv := c.client.KV()
	pairs, _, err := kv.List(c.prefix("peers"), nil)
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
