package client

import (
	"fmt"
	"math/rand"

	"github.com/hexablock/blockring"
	"github.com/hexablock/blockring/structs"
	"github.com/hexablock/blockring/utils"
	"github.com/hexablock/hexalog"
	chord "github.com/ipkg/go-chord"
)

// Config is the client configuration
type Config struct {
	ReapInterval  int      // interval at which outbound connections should be reaped.
	MaxIdle       int      // idle before a connection can be reaped.
	Peers         []string // initial set of peers.
	MaxSuccessors int
}

// DefaultConfig returns a sane client config
func DefaultConfig() *Config {
	return &Config{
		ReapInterval:  30,
		MaxIdle:       300,
		Peers:         []string{},
		MaxSuccessors: 3,
	}
}

// SetPeers parses a comma separated list of peers into a slice and sets the config.
func (conf *Config) SetPeers(peers string) {
	conf.Peers = utils.StringToSlice(peers, ",")
}

// Client is a client connecting to the core cluster
type Client struct {
	conf   *Config
	locate *blockring.LookupServiceClient
	rs     *blockring.BlockRing
	lr     *blockring.LogRing
}

// NewClient instantiates a new client
func NewClient(conf *Config) (*Client, error) {
	if len(conf.Peers) == 0 {
		return nil, fmt.Errorf("no peers found")
	}

	c := &Client{
		conf:   conf,
		locate: blockring.NewLookupServiceClient(conf.ReapInterval, conf.MaxIdle),
	}

	resp, err := c.locate.Negotiate(c.GetPeer())
	if err != nil {
		return nil, err
	}

	isucc := int(resp.Successors)
	if conf.MaxSuccessors == 0 || conf.MaxSuccessors > isucc {
		c.conf.MaxSuccessors = isucc
	}

	blkTrans := blockring.NewBlockNetTransportClient(conf.ReapInterval, conf.MaxIdle)
	c.rs = blockring.NewBlockRing(c, blkTrans, nil)

	logTrans := blockring.NewLogNetTransportClient(conf.ReapInterval, conf.MaxIdle)
	c.lr = blockring.NewLogRing(c, logTrans, nil)

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

func (client *Client) LookupHash(hash []byte, n int) (*chord.Vnode, []*chord.Vnode, error) {
	return client.locate.LookupHash(client.GetPeer(), hash, n)
}
func (client *Client) LookupKey(key []byte, n int) ([]byte, *chord.Vnode, []*chord.Vnode, error) {
	return client.locate.LookupKey(client.GetPeer(), key, n)
}
func (client *Client) LocateReplicatedHash(hash []byte, r int) ([]*structs.Location, error) {
	return client.locate.LocateReplicatedHash(client.GetPeer(), hash, r)
}
func (client *Client) LocateReplicatedKey(key []byte, r int) ([]*structs.Location, error) {
	return client.locate.LocateReplicatedKey(client.GetPeer(), key, r)
}

func (client *Client) NewTx(key []byte, opts hexalog.Options) (*hexalog.Tx, *hexalog.Meta, error) {
	return client.lr.NewTx(key, opts)
}
func (client *Client) ProposeTx(tx *hexalog.Tx, opts hexalog.Options) (*hexalog.Meta, error) {
	return client.lr.ProposeTx(tx, opts)
}

// SetBlock sets the given block on the ring with the configured replication factor
func (client *Client) SetBlock(block *structs.Block, opts ...blockring.RequestOptions) (*structs.Location, error) {
	return client.rs.SetBlock(block, opts...)
}

// GetBlock gets a block given the id.  It returns the first available block.
func (client *Client) GetBlock(id []byte, opts ...blockring.RequestOptions) (*structs.Location, *structs.Block, error) {
	return client.rs.GetBlock(id, opts...)
}
