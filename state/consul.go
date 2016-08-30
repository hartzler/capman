package state

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
)

// consul backend driver for storing external state of peers
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

func (c *consul) IsBootstrap() (*Bootstrap, error) {
	fmt.Println("Checking if initial quorum was ever reached...")
	var b Bootstrap
	kv := c.client.KV()
	pair, _, err := kv.Get(c.prefix("bootstrap"), nil)
	if err != nil {
		return nil, err
	}
	if pair == nil {
		return nil, nil
	}
	err = json.Unmarshal(pair.Value, &b)
	return &b, err
}

func (c *consul) SetBootstrap() (*Bootstrap, error) {
	fmt.Println("Marking cluster as boostrapped...")
	kv := c.client.KV()
	b := Bootstrap{time.Now()}
	bytes, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	_, err = kv.Put(&api.KVPair{Key: c.prefix("bootstrap"), Value: bytes}, nil)
	return &b, err
}

func (c *consul) Peers() (Peers, error) {
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
