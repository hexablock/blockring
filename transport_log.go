package blockring

import (
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

func (c *LogNetTransportClient) GetEntry(loc *structs.Location, hash []byte, opts structs.RequestOptions) (*structs.LogEntryBlock, *structs.Location, error) {
	conn, err := c.out.Get(loc.Vnode.Host)
	if err != nil {
		return nil, nil, err
	}

	req := &rpc.BlockRPCData{ID: hash, Options: &opts}
	resp, err := conn.LogRPC.GetEntryRPC(context.Background(), req)
	c.out.Return(conn)

	if err == nil {
		var l structs.LogEntryBlock
		if err = l.DecodeBlock(resp.Block); err == nil {
			return &l, resp.Location, nil
		}
	}

	return nil, resp.Location, err
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
	host string
	txl  *hexalog.HexaLog
}

func NewLogNetTransport(host string, txl *hexalog.HexaLog) *LogNetTransport {
	return &LogNetTransport{
		txl:  txl,
		host: host,
	}
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

func (t *LogNetTransport) GetEntryRPC(ctx context.Context, req *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {
	tx, loc, err := t.txl.GetEntry(req.ID)
	if err == nil {
		var blk *structs.Block
		if blk, err = tx.EncodeBlock(); err == nil {
			return &rpc.BlockRPCData{Block: blk, Location: loc}, nil
		}
	}
	return &rpc.BlockRPCData{}, err
}

func (t *LogNetTransport) ProposeEntryRPC(ctx context.Context, req *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {
	opts := *req.Options

	var entry structs.LogEntryBlock
	err := entry.UnmarshalBinary(req.Block.Data)
	if err == nil {
		err = t.txl.ProposeEntry(&entry, opts)
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
	remote LogTransport
}

func NewLogRingTransport(host string, txl *hexalog.HexaLog, remote LogTransport) *LogRingTransport {
	if remote == nil {
		return &LogRingTransport{host: host, remote: NewLogNetTransportClient(30, 180), txl: txl}
	}
	return &LogRingTransport{host: host, remote: remote, txl: txl}
}

func (lt *LogRingTransport) ProposeEntry(loc *structs.Location, tx *structs.LogEntryBlock, opts structs.RequestOptions) (*structs.Location, error) {
	if lt.host == loc.Vnode.Host {
		return &structs.Location{}, lt.txl.ProposeEntry(tx, opts)
	}
	return lt.remote.ProposeEntry(loc, tx, opts)
}

func (lt *LogRingTransport) NewEntry(loc *structs.Location, key []byte, opts structs.RequestOptions) (*structs.LogEntryBlock, *structs.Location, error) {
	if lt.host == loc.Vnode.Host {
		tx, err := lt.txl.NewEntry(key)
		return tx, &structs.Location{}, err
	}
	return lt.remote.NewEntry(loc, key, opts)
}

func (lt *LogRingTransport) GetEntry(loc *structs.Location, id []byte, opts structs.RequestOptions) (*structs.LogEntryBlock, *structs.Location, error) {
	if lt.host == loc.Vnode.Host {
		return lt.txl.GetEntry(id)
	}
	return lt.remote.GetEntry(loc, id, opts)
}

func (lt *LogRingTransport) CommitEntry(loc *structs.Location, tx *structs.LogEntryBlock, opts structs.RequestOptions) (*structs.Location, error) {
	if lt.host == loc.Vnode.Host {
		return &structs.Location{}, lt.txl.CommitEntry(tx, opts)
	}
	return lt.remote.CommitEntry(loc, tx, opts)
}
