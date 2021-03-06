package client

import (
	"fmt"
	"math/rand"

	"github.com/hexablock/blockring"
	"github.com/hexablock/blockring/structs"
	"github.com/hexablock/blockring/utils"
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
	conf     *Config
	locate   *blockring.LookupServiceClient
	logTrans blockring.LogTransport
	rs       *blockring.BlockRing
}

// NewClient instantiates a new client
func NewClient(conf *Config) *Client {
	return &Client{
		conf:     conf,
		locate:   blockring.NewLookupServiceClient(conf.ReapInterval, conf.MaxIdle),
		logTrans: blockring.NewLogNetTransportClient(conf.ReapInterval, conf.MaxIdle),
	}
}

// Configure negotiates ring parameters with peers and sets up the BlockRing.  This must be called
// before using the client.
func (client *Client) Configure() error {
	if len(client.conf.Peers) == 0 {
		return fmt.Errorf("no peers found")
	}

	resp, err := client.locate.Negotiate(client.GetPeer())
	if err != nil {
		return err
	}

	isucc := int(resp.Successors)
	if client.conf.MaxSuccessors == 0 || client.conf.MaxSuccessors > isucc {
		client.conf.MaxSuccessors = isucc
	}

	blkTrans := blockring.NewBlockNetTransportClient(client.conf.ReapInterval, client.conf.MaxIdle)
	client.rs = blockring.NewBlockRing(client, int(resp.Successors), blkTrans, client.logTrans, nil)

	return nil
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

// NewEntry retrieves a new LogEntryBlock that would be next in the chain.
func (client *Client) NewEntry(key []byte, opts structs.RequestOptions) (*structs.LogEntryBlock, *structs.Location, error) {
	return client.rs.NewEntry(key, opts)
}

// ProposeEntry submits a ProposeEntry request to the network.
func (client *Client) ProposeEntry(tx *structs.LogEntryBlock, opts structs.RequestOptions) error {
	//return client.rs.ProposeEntry(tx, opts)

	locs, err := client.LocateReplicatedKey(tx.Key, int(opts.PeerSetSize))
	if err != nil {
		return err
	}

	return client.logTrans.ProposeEntry(locs[0], tx, opts)
}

// GetLogBlock gets the first LogBlock of the given key
func (client *Client) GetLogBlock(key []byte, opts structs.RequestOptions) (*structs.Location, *structs.LogBlock, error) {
	return client.rs.GetLogBlock(key, opts)
}

// SetBlock sets the given block on the ring with the configured replication factor
func (client *Client) SetBlock(block *structs.Block, opts ...structs.RequestOptions) (*structs.Location, error) {
	return client.rs.SetBlock(block, opts...)
}

// GetBlock gets a block given the id.  It returns the first available block.
func (client *Client) GetBlock(id []byte, opts ...structs.RequestOptions) (*structs.Location, *structs.Block, error) {
	return client.rs.GetBlock(id, opts...)
}
