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

func (c *LogNetTransportClient) ProposeEntry(loc *structs.Location, tx *hexalog.Entry, opts hexalog.Options) (*hexalog.Meta, error) {
	conn, err := c.out.Get(loc.Vnode.Host)
	if err != nil {
		return nil, err
	}

	req := &rpc.LogRPCData{Entry: tx, Options: &opts}
	resp, err := conn.LogRPC.ProposeEntryRPC(context.Background(), req)
	c.out.Return(conn)
	if err == nil {
		return resp.Meta, nil
	}
	return nil, err
}

func (c *LogNetTransportClient) NewEntry(loc *structs.Location, key []byte, opts hexalog.Options) (*hexalog.Entry, *hexalog.Meta, error) {
	conn, err := c.out.Get(loc.Vnode.Host)
	if err != nil {
		return nil, nil, err
	}

	req := &rpc.LogRPCData{Options: &opts, Id: key}
	resp, err := conn.LogRPC.NewEntryRPC(context.Background(), req)
	c.out.Return(conn)

	return resp.Entry, resp.Meta, err
}

func (c *LogNetTransportClient) GetEntry(loc *structs.Location, hash []byte, opts hexalog.Options) (*hexalog.Entry, *hexalog.Meta, error) {
	conn, err := c.out.Get(loc.Vnode.Host)
	if err != nil {
		return nil, nil, err
	}

	req := &rpc.LogRPCData{Id: hash, Options: &opts}
	resp, err := conn.LogRPC.GetEntryRPC(context.Background(), req)
	c.out.Return(conn)

	return resp.Entry, resp.Meta, err
}

func (c *LogNetTransportClient) CommitEntry(loc *structs.Location, tx *hexalog.Entry, opts hexalog.Options) (*hexalog.Meta, error) {
	conn, err := c.out.Get(loc.Vnode.Host)
	if err != nil {
		return nil, err
	}

	req := &rpc.LogRPCData{Entry: tx, Options: &opts}
	resp, err := conn.LogRPC.CommitEntryRPC(context.Background(), req)
	c.out.Return(conn)
	if err == nil {
		return resp.Meta, nil
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

func (t *LogNetTransport) NewEntryRPC(ctx context.Context, req *rpc.LogRPCData) (*rpc.LogRPCData, error) {
	tx, err := t.txl.NewEntry(req.Id)
	if err == nil {
		return &rpc.LogRPCData{Entry: tx}, nil
	}
	return &rpc.LogRPCData{}, err
}

func (t *LogNetTransport) GetEntryRPC(ctx context.Context, req *rpc.LogRPCData) (*rpc.LogRPCData, error) {
	tx, meta, err := t.txl.GetEntry(req.Id)
	if err == nil {
		return &rpc.LogRPCData{Entry: tx, Meta: meta}, nil
	}
	return &rpc.LogRPCData{}, err
}

func (t *LogNetTransport) ProposeEntryRPC(ctx context.Context, req *rpc.LogRPCData) (*rpc.LogRPCData, error) {
	opts := *req.Options
	// if opts.SourceHost == "" {
	// 	opts.SourceHost = t.host
	// }

	if err := t.txl.ProposeEntry(req.Entry, opts); err != nil {
		return &rpc.LogRPCData{}, err
	}
	return &rpc.LogRPCData{}, nil
}

func (t *LogNetTransport) CommitEntryRPC(ctx context.Context, req *rpc.LogRPCData) (*rpc.LogRPCData, error) {
	opts := *req.Options
	// if opts.SourceHost == "" {
	// 	opts.SourceHost = t.host
	// }
	if err := t.txl.CommitEntry(req.Entry, opts); err != nil {
		return &rpc.LogRPCData{}, err
	}
	return &rpc.LogRPCData{}, nil
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

func (lt *LogRingTransport) ProposeEntry(loc *structs.Location, tx *hexalog.Entry, opts hexalog.Options) (*hexalog.Meta, error) {
	if lt.host == loc.Vnode.Host {
		return &hexalog.Meta{}, lt.txl.ProposeEntry(tx, opts)
	}
	return lt.remote.ProposeEntry(loc, tx, opts)
}

func (lt *LogRingTransport) NewEntry(loc *structs.Location, key []byte, opts hexalog.Options) (*hexalog.Entry, *hexalog.Meta, error) {
	if lt.host == loc.Vnode.Host {
		tx, err := lt.txl.NewEntry(key)
		return tx, &hexalog.Meta{}, err
	}
	return lt.remote.NewEntry(loc, key, opts)
}

func (lt *LogRingTransport) GetEntry(loc *structs.Location, id []byte, opts hexalog.Options) (*hexalog.Entry, *hexalog.Meta, error) {
	if lt.host == loc.Vnode.Host {
		return lt.txl.GetEntry(id)
	}
	return lt.remote.GetEntry(loc, id, opts)
}

func (lt *LogRingTransport) CommitEntry(loc *structs.Location, tx *hexalog.Entry, opts hexalog.Options) (*hexalog.Meta, error) {
	if lt.host == loc.Vnode.Host {
		return &hexalog.Meta{}, lt.txl.CommitEntry(tx, opts)
	}
	return lt.remote.CommitEntry(loc, tx, opts)
}
