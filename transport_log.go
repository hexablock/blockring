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
func (c *LogNetTransportClient) ProposeEntry(loc *structs.Location, tx *structs.LogEntryBlock, opts structs.RequestOptions) error {

	blk, err := tx.EncodeBlock()
	if err != nil {
		return err
	}

	conn, err := c.out.Get(loc.Vnode.Host)
	if err != nil {
		return err
	}

	req := &rpc.BlockRPCData{Block: blk, Options: &opts}
	_, err = conn.LogRPC.ProposeEntryRPC(context.Background(), req)
	c.out.Return(conn)
	//if err == nil {
	//return  nil
	//}
	return err
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
func (c *LogNetTransportClient) TransferLogBlock(key []byte, src, dst *structs.Location) error {

	conn, err := c.out.Get(dst.Vnode.Host)
	if err != nil {
		//return nil, err
		return err
	}

	//req := &rpc.BlockRPCData{ID: key, Options: &opts}
	req := &rpc.RelocateRPCData{ID: key, Source: src, Destination: dst}
	_, err = conn.LogRPC.TransferLogBlockRPC(context.Background(), req)
	c.out.Return(conn)

	return err
}

// CommitEntry makes a CommitEntry request to the Location
func (c *LogNetTransportClient) CommitEntry(loc *structs.Location, tx *structs.LogEntryBlock, opts structs.RequestOptions) error {

	blk, err := tx.EncodeBlock()
	if err != nil {
		return err
	}

	conn, err := c.out.Get(loc.Vnode.Host)
	if err != nil {
		return err
	}

	req := &rpc.BlockRPCData{Block: blk, Options: &opts}
	_, err = conn.LogRPC.CommitEntryRPC(context.Background(), req)
	c.out.Return(conn)
	return err
}

// LogNetTransport is the network transport for the consensus log.
type LogNetTransport struct {
	host     string
	txl      *hexalog.HexaLog
	bs       hexalog.BlockStore
	relocate chan *rpc.RelocateRPCData
}

// NewLogNetTransport instantiates a new log network transport with the host, HexaLog and BlockStore
func NewLogNetTransport(host string, txl *hexalog.HexaLog, bs hexalog.BlockStore) *LogNetTransport {
	return &LogNetTransport{
		host: host,
		txl:  txl,
		bs:   bs,
	}
}

// Register registers a channel where incoming blocks are sent for processing.
func (t *LogNetTransport) Register(ch chan *rpc.RelocateRPCData) {
	t.relocate = ch
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
func (t *LogNetTransport) TransferLogBlockRPC(ctx context.Context, req *rpc.RelocateRPCData) (*rpc.RelocateRPCData, error) {
	req.Block = &structs.Block{Type: structs.BlockType_LOG}
	t.relocate <- req

	return &rpc.RelocateRPCData{}, nil
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

// GetEntryRPC serves a GetEntry request.
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

// NewEntryRPC services a NewEntry request
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

// ProposeEntryRPC serves a ProposeEntry request
func (t *LogNetTransport) ProposeEntryRPC(ctx context.Context, req *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {
	opts := *req.Options

	var entry structs.LogEntryBlock
	err := entry.UnmarshalBinary(req.Block.Data)
	if err == nil {
		//var ballot *hexalog.Ballot
		if _, err = t.txl.ProposeEntry(&entry, opts); err == nil {
			//log.Printf("Waiting on ballot: %p", ballot)
			//err = ballot.Wait()
		}
	}
	//log.Println("Returning from ProposeEntryRPC")

	return &rpc.BlockRPCData{}, err
}

// CommitEntryRPC serves a CommitEntry request
func (t *LogNetTransport) CommitEntryRPC(ctx context.Context, req *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {
	opts := *req.Options
	var entry structs.LogEntryBlock
	err := entry.UnmarshalBinary(req.Block.Data)
	if err == nil {
		err = t.txl.CommitEntry(&entry, opts)
	}

	return &rpc.BlockRPCData{}, err
}

// LogRingTransport is a transport to make log requests around the ring based on location.
type LogRingTransport struct {
	host   string
	txl    *hexalog.HexaLog
	bs     hexalog.BlockStore
	remote LogTransport
}

// NewLogRingTransport instantiates a new LogRingTransport to make location based calls around the ring
func NewLogRingTransport(host string, txl *hexalog.HexaLog, bs hexalog.BlockStore, remote LogTransport) *LogRingTransport {
	if remote == nil {
		return &LogRingTransport{host: host, remote: NewLogNetTransportClient(30, 180), txl: txl, bs: bs}
	}
	return &LogRingTransport{host: host, remote: remote, txl: txl, bs: bs}
}

// GetLogBlock retrieves a LogBlock from a local or remote Location
func (lt *LogRingTransport) GetLogBlock(loc *structs.Location, key []byte, opts structs.RequestOptions) (*structs.LogBlock, *structs.Location, error) {
	if lt.host == loc.Vnode.Host {
		lb, err := lt.bs.Get(key)
		return lb, &structs.Location{}, err
	}
	return lt.remote.GetLogBlock(loc, key, opts)
}

// ProposeEntry proposes a LogEntryBlock to a local or remote Location
func (lt *LogRingTransport) ProposeEntry(loc *structs.Location, tx *structs.LogEntryBlock, opts structs.RequestOptions) error {
	if lt.host == loc.Vnode.Host {
		_, err := lt.txl.ProposeEntry(tx, opts)
		return err
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

// NewEntry creates a new LogEntryBlock from the provided Location
func (lt *LogRingTransport) NewEntry(loc *structs.Location, key []byte, opts structs.RequestOptions) (*structs.LogEntryBlock, *structs.Location, error) {
	if lt.host == loc.Vnode.Host {
		tx, err := lt.txl.NewEntry(key)
		return tx, &structs.Location{}, err
	}
	return lt.remote.NewEntry(loc, key, opts)
}

// CommitEntry commits the LogEntryBlock to the specified Location
func (lt *LogRingTransport) CommitEntry(loc *structs.Location, tx *structs.LogEntryBlock, opts structs.RequestOptions) error {
	if lt.host == loc.Vnode.Host {
		return lt.txl.CommitEntry(tx, opts)
	}
	return lt.remote.CommitEntry(loc, tx, opts)
}

// TransferLogBlock makes a transfer request to the destination.  It returns an error if a local
// transfer is initiated otherwise the remote transport error is returned.
func (lt *LogRingTransport) TransferLogBlock(key []byte, src, dst *structs.Location) error {
	if lt.host == dst.Vnode.Host {
		return errNoLocalTransfer
	}
	return lt.remote.TransferLogBlock(key, src, dst)
}
