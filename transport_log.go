package blockring

import (
	"fmt"

	"golang.org/x/net/context"

	"github.com/hexablock/blockring/pool"
	"github.com/hexablock/blockring/rpc"
	"github.com/hexablock/blockring/structs"
	"github.com/hexablock/hexalog"
)

type LogNetTransportClient struct {
	out *pool.OutConnPool
}

func NewLogNetTransportClient(reapInterval, maxIdle int) *LogNetTransportClient {
	return &LogNetTransportClient{out: pool.NewOutConnPool(reapInterval, maxIdle)}
}

// ProposeEntry makes an ProposeEntry rpc call to a location.
func (c *LogNetTransportClient) ProposeEntry(loc *structs.Location, tx *structs.LogEntryBlock, opts structs.RequestOptions) (*structs.Location, error) {

	blk, err := tx.EncodeBlock()
	if err != nil {
		return nil, err
	}

	conn, err := c.out.Get(loc.Vnode.Host)
	if err != nil {
		return nil, err
	}

	req := &rpc.BlockRPCData{Block: blk, Options: &opts}
	resp, err := conn.LogRPC.ProposeEntryRPC(context.Background(), req)
	c.out.Return(conn)
	if err == nil {
		return resp.Location, nil
	}
	return nil, err
}

// NewEntry makes a new a entry request to Location for key
func (c *LogNetTransportClient) NewEntry(loc *structs.Location, key []byte, opts structs.RequestOptions) (*structs.LogEntryBlock, *structs.Location, error) {
	conn, err := c.out.Get(loc.Vnode.Host)
	if err != nil {
		return nil, nil, err
	}

	req := &rpc.BlockRPCData{Options: &opts, ID: key}
	resp, err := conn.LogRPC.NewEntryRPC(context.Background(), req)
	c.out.Return(conn)

	if err == nil {
		var l structs.LogEntryBlock
		if err = l.DecodeBlock(resp.Block); err == nil {
			return &l, resp.Location, nil
		}
	}

	return nil, resp.Location, err
}

// GetEntry makes a GetEntry rpc call to the location with the id.
func (c *LogNetTransportClient) GetEntry(loc *structs.Location, id []byte, opts structs.RequestOptions) (*structs.LogEntryBlock, *structs.Location, error) {
	conn, err := c.out.Get(loc.Vnode.Host)
	if err != nil {
		return nil, nil, err
	}

	req := &rpc.BlockRPCData{Options: &opts, ID: id}
	resp, err := conn.LogRPC.GetEntryRPC(context.Background(), req)
	c.out.Return(conn)

	if err == nil {
		var l structs.LogEntryBlock
		err = l.DecodeBlock(resp.Block)
		return &l, resp.Location, err
	}
	return nil, nil, err
}

// GetLogBlock gets a LogBlock from the provided location.
func (c *LogNetTransportClient) GetLogBlock(loc *structs.Location, key []byte, opts structs.RequestOptions) (*structs.LogBlock, *structs.Location, error) {
	conn, err := c.out.Get(loc.Vnode.Host)
	if err != nil {
		return nil, nil, err
	}

	req := &rpc.BlockRPCData{ID: key, Options: &opts}
	resp, err := conn.LogRPC.GetLogBlockRPC(context.Background(), req)
	c.out.Return(conn)

	if err == nil {
		var lb structs.LogBlock
		if err = lb.DecodeBlock(resp.Block); err == nil {
			return &lb, resp.Location, nil
		}
	}

	return nil, nil, err
}

// TransferLogBlock submits a transfer request to the location for a LogBlock by the key
func (c *LogNetTransportClient) TransferLogBlock(loc *structs.Location, key []byte, opts structs.RequestOptions) (*structs.Location, error) {
	conn, err := c.out.Get(loc.Vnode.Host)
	if err != nil {
		return nil, err
	}

	req := &rpc.BlockRPCData{ID: key, Options: &opts}
	resp, err := conn.LogRPC.TransferLogBlockRPC(context.Background(), req)
	c.out.Return(conn)

	if err == nil {
		return resp.Location, nil
	}

	return nil, err
}

func (c *LogNetTransportClient) CommitEntry(loc *structs.Location, tx *structs.LogEntryBlock, opts structs.RequestOptions) (*structs.Location, error) {

	blk, err := tx.EncodeBlock()
	if err != nil {
		return nil, err
	}

	conn, err := c.out.Get(loc.Vnode.Host)
	if err != nil {
		return nil, err
	}

	req := &rpc.BlockRPCData{Block: blk, Options: &opts}
	resp, err := conn.LogRPC.CommitEntryRPC(context.Background(), req)
	c.out.Return(conn)
	if err == nil {
		return resp.Location, nil
	}
	return nil, err
}

type LogNetTransport struct {
	host        string
	txl         *hexalog.HexaLog
	bs          hexalog.BlockStore
	inLogBlocks chan *rpc.BlockRPCData
}

func NewLogNetTransport(host string, txl *hexalog.HexaLog, bs hexalog.BlockStore) *LogNetTransport {
	return &LogNetTransport{
		txl:  txl,
		bs:   bs,
		host: host,
	}
}

// Register registers a channel where incoming blocks are sent for processing.
func (t *LogNetTransport) Register(ch chan *rpc.BlockRPCData) {
	t.inLogBlocks = ch
}

// GetLogBlockRPC serves a GetLogBlock request.
func (t *LogNetTransport) GetLogBlockRPC(ctx context.Context, req *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {
	resp := &rpc.BlockRPCData{}

	blk, err := t.bs.Get(req.ID)
	if err == nil {
		resp.Block, err = blk.EncodeBlock()
	}
	return resp, err
}

// TransferLogBlockRPC submits a LogBlock transfer request to the local input block channel.
func (t *LogNetTransport) TransferLogBlockRPC(ctx context.Context, req *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {
	resp := &rpc.BlockRPCData{}

	req.Block = &structs.Block{Type: structs.BlockType_LOG}
	t.inLogBlocks <- req

	return resp, nil
}

func (t *LogNetTransport) ReleaseLogBlockRPC(ctx context.Context, req *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {
	resp := &rpc.BlockRPCData{}
	//
	// blk, err := t.bs.Get(req.ID)
	// if err == nil {
	// 	resp.Block, err = blk.EncodeBlock()
	// }
	// return resp, err
	return resp, fmt.Errorf("TBI")
}

func (t *LogNetTransport) GetEntryRPC(ctx context.Context, req *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {
	entry, loc, err := t.txl.GetEntry(req.ID)
	if err == nil {
		var blk *structs.Block
		if blk, err = entry.EncodeBlock(); err == nil {
			return &rpc.BlockRPCData{Block: blk, Location: loc}, nil
		}
	}
	return &rpc.BlockRPCData{}, err
}

func (t *LogNetTransport) NewEntryRPC(ctx context.Context, req *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {
	tx, err := t.txl.NewEntry(req.ID)
	if err == nil {
		var blk *structs.Block
		if blk, err = tx.EncodeBlock(); err == nil {
			return &rpc.BlockRPCData{Block: blk}, nil
		}

	}
	return &rpc.BlockRPCData{}, err
}

func (t *LogNetTransport) ProposeEntryRPC(ctx context.Context, req *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {
	opts := *req.Options

	var entry structs.LogEntryBlock
	err := entry.UnmarshalBinary(req.Block.Data)
	if err == nil {
		_, err = t.txl.ProposeEntry(&entry, opts)
	}

	return &rpc.BlockRPCData{}, err
}

func (t *LogNetTransport) CommitEntryRPC(ctx context.Context, req *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {
	opts := *req.Options
	var entry structs.LogEntryBlock
	err := entry.UnmarshalBinary(req.Block.Data)
	if err == nil {
		err = t.txl.CommitEntry(&entry, opts)
	}

	return &rpc.BlockRPCData{}, err
}

type LogRingTransport struct {
	host   string
	txl    *hexalog.HexaLog
	bs     hexalog.BlockStore
	remote LogTransport
}

func NewLogRingTransport(host string, txl *hexalog.HexaLog, bs hexalog.BlockStore, remote LogTransport) *LogRingTransport {
	if remote == nil {
		return &LogRingTransport{host: host, remote: NewLogNetTransportClient(30, 180), txl: txl, bs: bs}
	}
	return &LogRingTransport{host: host, remote: remote, txl: txl, bs: bs}
}

func (lt *LogRingTransport) GetLogBlock(loc *structs.Location, key []byte, opts structs.RequestOptions) (*structs.LogBlock, *structs.Location, error) {
	if lt.host == loc.Vnode.Host {
		lb, err := lt.bs.Get(key)
		return lb, &structs.Location{}, err
	}
	return lt.remote.GetLogBlock(loc, key, opts)
}

func (lt *LogRingTransport) TransferLogBlock(loc *structs.Location, key []byte, opts structs.RequestOptions) (*structs.Location, error) {
	if lt.host == loc.Vnode.Host {
		return nil, errNoLocalTransfer
	}
	return lt.remote.TransferLogBlock(loc, key, opts)
}

func (lt *LogRingTransport) ProposeEntry(loc *structs.Location, tx *structs.LogEntryBlock, opts structs.RequestOptions) (*structs.Location, error) {
	if lt.host == loc.Vnode.Host {
		_, err := lt.txl.ProposeEntry(tx, opts)
		return &structs.Location{}, err
	}
	return lt.remote.ProposeEntry(loc, tx, opts)
}

// GetEntry gets a log entry from a location returning the entry and the location it was retreived from.
func (lt *LogRingTransport) GetEntry(loc *structs.Location, id []byte, opts structs.RequestOptions) (*structs.LogEntryBlock, *structs.Location, error) {

	if lt.host == loc.Vnode.Host {
		return lt.txl.GetEntry(id)
	}
	return lt.remote.GetEntry(loc, id, opts)
}

func (lt *LogRingTransport) NewEntry(loc *structs.Location, key []byte, opts structs.RequestOptions) (*structs.LogEntryBlock, *structs.Location, error) {
	if lt.host == loc.Vnode.Host {
		tx, err := lt.txl.NewEntry(key)
		return tx, &structs.Location{}, err
	}
	return lt.remote.NewEntry(loc, key, opts)
}

func (lt *LogRingTransport) CommitEntry(loc *structs.Location, tx *structs.LogEntryBlock, opts structs.RequestOptions) (*structs.Location, error) {
	if lt.host == loc.Vnode.Host {
		return &structs.Location{}, lt.txl.CommitEntry(tx, opts)
	}
	return lt.remote.CommitEntry(loc, tx, opts)
}
