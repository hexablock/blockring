package client

import (
	"fmt"
	"math/rand"

	"github.com/hexablock/blockring"
	"github.com/hexablock/blockring/structs"
	"github.com/hexablock/blockring/utils"
)

// Config is the client configuration
type Config struct {
	ReapInterval int      // interval at which outbound connections should be reaped.
	MaxIdle      int      // idle before a connection can be reaped.
	Peers        []string // initial set of peers.
}

// DefaultConfig returns a sane client config
func DefaultConfig() *Config {
	return &Config{
		ReapInterval: 30,
		MaxIdle:      300,
		Peers:        []string{},
	}
}

// SetPeers parses a comma separated list of peers into a slice and sets the config.
func (conf *Config) SetPeers(peers string) {
	conf.Peers = utils.StringToSlice(peers, ",")
}

// Client is a client connecting to the core cluster
type Client struct {
	conf *Config

	store  *blockring.NetTransportClient
	locate *blockring.LookupServiceClient
}

// NewClient instantiates a new client
func NewClient(conf *Config) (*Client, error) {
	if len(conf.Peers) == 0 {
		return nil, fmt.Errorf("no peers found")
	}

	c := &Client{
		conf:   conf,
		store:  blockring.NewNetTransportClient(conf.ReapInterval, conf.MaxIdle),
		locate: blockring.NewLookupServiceClient(conf.ReapInterval, conf.MaxIdle),
	}

	return c, nil
}

// GetPeer returns a single random peer
func (client *Client) GetPeer() string {
	if len(client.conf.Peers) == 0 {
		return ""
	}

	i := rand.Int() % len(client.conf.Peers)
	return client.conf.Peers[i]
}

// SetBlock sets the given block on the ring with the configured replication factor
func (client *Client) SetBlock(block *structs.Block, opts ...RequestOptions) error {
	id := block.ID()

	_, succs, err := client.locate.LookupHash(client.GetPeer(), id, 1)
	//locs, err := client.locate.LocateBlock(client.conf.Peers[0], id)
	if err != nil {
		return err
	}

	loc := &structs.Location{Id: id}

	for _, succ := range succs {
		loc.Vnode = succ
		er := client.store.SetBlock(loc, block)
		if er != nil {
			err = er
		}
	}

	return err
}

// GetBlock gets a block given the id.  It returns the first available block.
func (client *Client) GetBlock(id []byte, opts ...RequestOptions) (*structs.Location, *structs.Block, error) {

	_, succs, err := client.locate.LookupHash(client.GetPeer(), id, 3)
	if err != nil {
		return nil, nil, err
	}

	/*locs, err := client.locate.LocateBlock(client.conf.Peers[0], id)
	if err != nil {
		return nil, err
	}*/

	loc := &structs.Location{Id: id}
	for _, succ := range succs {
		loc.Vnode = succ
		block, er := client.store.GetBlock(loc, id)
		if er == nil {
			return loc, block, nil
		}
		err = er
	}

	return loc, nil, err
}

// GetRootBlock gets a root block with the given id
func (client *Client) GetRootBlock(id []byte, opts ...RequestOptions) (*structs.Location, *structs.RootBlock, error) {
	loc, block, err := client.GetBlock(id, opts...)
	if err == nil {
		var rb structs.RootBlock
		err = rb.DecodeBlock(block)
		return loc, &rb, err
	}
	return loc, nil, err
}
