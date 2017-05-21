package blockring

import (
	"golang.org/x/net/context"

	"github.com/hexablock/blockring/pool"
	"github.com/hexablock/blockring/rpc"
	"github.com/hexablock/blockring/structs"
	"github.com/hexablock/txlog"
)

type LogNetTransportClient struct {
	out *pool.OutConnPool
}

func NewLogNetTransportClient(reapInterval, maxIdle int) *LogNetTransportClient {
	return &LogNetTransportClient{out: pool.NewOutConnPool(reapInterval, maxIdle)}
}

func (c *LogNetTransportClient) ProposeTx(loc *structs.Location, tx *txlog.Tx, opts txlog.Options) (*txlog.Meta, error) {
	conn, err := c.out.Get(loc.Vnode.Host)
	if err != nil {
		return nil, err
	}

	req := &rpc.LogRPCData{Tx: tx, Options: &opts}
	resp, err := conn.LogRPC.ProposeTxRPC(context.Background(), req)
	c.out.Return(conn)
	if err == nil {
		return resp.Meta, nil
	}
	return nil, err
}

func (c *LogNetTransportClient) NewTx(loc *structs.Location, key []byte, opts txlog.Options) (*txlog.Tx, *txlog.Meta, error) {
	conn, err := c.out.Get(loc.Vnode.Host)
	if err != nil {
		return nil, nil, err
	}

	req := &rpc.LogRPCData{Options: &opts, Id: key}
	resp, err := conn.LogRPC.NewTxRPC(context.Background(), req)
	c.out.Return(conn)

	return resp.Tx, resp.Meta, err
}

func (c *LogNetTransportClient) GetTx(loc *structs.Location, hash []byte, opts txlog.Options) (*txlog.Tx, *txlog.Meta, error) {
	conn, err := c.out.Get(loc.Vnode.Host)
	if err != nil {
		return nil, nil, err
	}

	req := &rpc.LogRPCData{Id: hash, Options: &opts}
	resp, err := conn.LogRPC.GetTxRPC(context.Background(), req)
	c.out.Return(conn)

	return resp.Tx, resp.Meta, err
}

func (c *LogNetTransportClient) CommitTx(loc *structs.Location, tx *txlog.Tx, opts txlog.Options) (*txlog.Meta, error) {
	conn, err := c.out.Get(loc.Vnode.Host)
	if err != nil {
		return nil, err
	}

	req := &rpc.LogRPCData{Tx: tx, Options: &opts}
	resp, err := conn.LogRPC.CommitTxRPC(context.Background(), req)
	c.out.Return(conn)
	if err == nil {
		return resp.Meta, nil
	}
	return nil, err
}

type LogNetTransport struct {
	host string
	txl  *txlog.TxLog
}

func NewLogNetTransport(host string, txl *txlog.TxLog) *LogNetTransport {
	return &LogNetTransport{
		txl:  txl,
		host: host,
	}
}

func (t *LogNetTransport) NewTxRPC(ctx context.Context, req *rpc.LogRPCData) (*rpc.LogRPCData, error) {
	tx, err := t.txl.NewTx(req.Id)
	if err == nil {
		return &rpc.LogRPCData{Tx: tx}, nil
	}
	return &rpc.LogRPCData{}, err
}

func (t *LogNetTransport) GetTxRPC(ctx context.Context, req *rpc.LogRPCData) (*rpc.LogRPCData, error) {
	tx, meta, err := t.txl.GetTx(req.Id)
	if err == nil {
		return &rpc.LogRPCData{Tx: tx, Meta: meta}, nil
	}
	return &rpc.LogRPCData{}, err
}

func (t *LogNetTransport) ProposeTxRPC(ctx context.Context, req *rpc.LogRPCData) (*rpc.LogRPCData, error) {
	opts := *req.Options
	// if opts.SourceHost == "" {
	// 	opts.SourceHost = t.host
	// }

	if err := t.txl.ProposeTx(req.Tx, opts); err != nil {
		return &rpc.LogRPCData{}, err
	}
	return &rpc.LogRPCData{}, nil
}

func (t *LogNetTransport) CommitTxRPC(ctx context.Context, req *rpc.LogRPCData) (*rpc.LogRPCData, error) {
	opts := *req.Options
	// if opts.SourceHost == "" {
	// 	opts.SourceHost = t.host
	// }
	if err := t.txl.CommitTx(req.Tx, opts); err != nil {
		return &rpc.LogRPCData{}, err
	}
	return &rpc.LogRPCData{}, nil
}

type LogRingTransport struct {
	host   string
	txl    *txlog.TxLog
	remote LogTransport
}

func NewLogRingTransport(host string, txl *txlog.TxLog, remote LogTransport) *LogRingTransport {
	if remote == nil {
		return &LogRingTransport{host: host, remote: NewLogNetTransportClient(30, 180), txl: txl}
	}
	return &LogRingTransport{host: host, remote: remote, txl: txl}
}

func (lt *LogRingTransport) ProposeTx(loc *structs.Location, tx *txlog.Tx, opts txlog.Options) (*txlog.Meta, error) {
	if lt.host == loc.Vnode.Host {
		return &txlog.Meta{}, lt.txl.ProposeTx(tx, opts)
	}
	return lt.remote.ProposeTx(loc, tx, opts)
}

func (lt *LogRingTransport) NewTx(loc *structs.Location, key []byte, opts txlog.Options) (*txlog.Tx, *txlog.Meta, error) {
	if lt.host == loc.Vnode.Host {
		tx, err := lt.txl.NewTx(key)
		return tx, &txlog.Meta{}, err
	}
	return lt.remote.NewTx(loc, key, opts)
}

func (lt *LogRingTransport) GetTx(loc *structs.Location, id []byte, opts txlog.Options) (*txlog.Tx, *txlog.Meta, error) {
	if lt.host == loc.Vnode.Host {
		return lt.txl.GetTx(id)
	}
	return lt.remote.GetTx(loc, id, opts)
}

func (lt *LogRingTransport) CommitTx(loc *structs.Location, tx *txlog.Tx, opts txlog.Options) (*txlog.Meta, error) {
	if lt.host == loc.Vnode.Host {
		return &txlog.Meta{}, lt.txl.CommitTx(tx, opts)
	}
	return lt.remote.CommitTx(loc, tx, opts)
}
