package blockring

import (
	"errors"

	"golang.org/x/net/context"

	"github.com/hexablock/blockring/pool"
	"github.com/hexablock/blockring/rpc"
	"github.com/hexablock/blockring/store"
	"github.com/hexablock/blockring/structs"
)

var (
	errNoLocalTransfer = errors.New("local transfers not allowed")
)

type BlockNetTransportClient struct {
	out *pool.OutConnPool
}

func NewBlockNetTransportClient(reapInterval, maxIdle int) *BlockNetTransportClient {
	return &BlockNetTransportClient{out: pool.NewOutConnPool(reapInterval, maxIdle)}
}

func (s *BlockNetTransportClient) GetBlock(loc *structs.Location, id []byte) (*structs.Block, error) {
	conn, err := s.out.Get(loc.Vnode.Host)
	if err != nil {
		return nil, err
	}
	defer s.out.Return(conn)

	req := &rpc.BlockRPCData{ID: id}
	resp, err := conn.BlockRPC.GetBlockRPC(context.Background(), req)
	if err == nil {
		return resp.Block, nil
	}
	return nil, err
}

func (s *BlockNetTransportClient) SetBlock(loc *structs.Location, block *structs.Block) error {
	conn, err := s.out.Get(loc.Vnode.Host)
	if err != nil {
		return err
	}

	req := &rpc.BlockRPCData{Block: block}
	_, err = conn.BlockRPC.SetBlockRPC(context.Background(), req)

	s.out.Return(conn)

	return err
}

// TransferBlock submits a transfer request to the location for a Block by its id
func (s *BlockNetTransportClient) TransferBlock(id []byte, src, dst *structs.Location) error {
	conn, err := s.out.Get(dst.Vnode.Host)
	if err != nil {
		return err
	}

	// TODO: use only block ids
	req := &rpc.RelocateRPCData{ID: id, Source: src, Destination: dst}
	_, err = conn.BlockRPC.TransferBlockRPC(context.Background(), req)

	s.out.Return(conn)

	return err
}

func (s *BlockNetTransportClient) ReleaseBlock(loc *structs.Location, id []byte) error {
	conn, err := s.out.Get(loc.Vnode.Host)
	if err != nil {
		return err
	}

	req := &rpc.BlockRPCData{ID: id}
	_, err = conn.BlockRPC.ReleaseBlockRPC(context.Background(), req)

	s.out.Return(conn)

	return err
}

type BlockNetTransport struct {
	st store.BlockStore
	// potential inbound blocks
	inBlocks chan *rpc.RelocateRPCData
}

func NewBlockNetTransport(bs store.BlockStore) *BlockNetTransport {
	return &BlockNetTransport{st: bs}
}

// Register registers a channel where incoming blocks are sent for processing.
func (s *BlockNetTransport) Register(ch chan *rpc.RelocateRPCData) {
	s.inBlocks = ch
}

func (s *BlockNetTransport) GetBlockRPC(ctx context.Context, in *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {

	blk, err := s.st.GetBlock(in.ID)
	if err != nil {
		return nil, err
	}

	return &rpc.BlockRPCData{Block: blk}, nil
}

func (s *BlockNetTransport) SetBlockRPC(ctx context.Context, in *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {
	err := s.st.SetBlock(in.Block)
	return &rpc.BlockRPCData{}, err
}

func (s *BlockNetTransport) TransferBlockRPC(ctx context.Context, in *rpc.RelocateRPCData) (*rpc.RelocateRPCData, error) {
	s.inBlocks <- in
	return &rpc.RelocateRPCData{}, nil
}

func (s *BlockNetTransport) ReleaseBlockRPC(ctx context.Context, in *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {
	err := s.st.ReleaseBlock(in.ID)
	return &rpc.BlockRPCData{}, err
}

// BlockRingTransport allows to make block requests based on Location around the ring.  It appropriately
// makes calls to the local or remote locations
type BlockRingTransport struct {
	host   string
	local  store.BlockStore
	remote BlockTransport
}

func NewBlockRingTransport(hostname string, local store.BlockStore, remote BlockTransport) *BlockRingTransport {
	//remote.RegisterStore(local)
	st := &BlockRingTransport{host: hostname, local: local, remote: remote}
	if st.remote == nil {
		st.remote = NewBlockNetTransportClient(30, 180)
	}
	return st
}

func (t *BlockRingTransport) GetBlock(loc *structs.Location, id []byte) (*structs.Block, error) {
	if loc.Vnode.Host == t.host {
		return t.local.GetBlock(id)
	}

	return t.remote.GetBlock(loc, id)
}

func (t *BlockRingTransport) SetBlock(loc *structs.Location, block *structs.Block) error {
	if loc.Vnode.Host == t.host {
		return t.local.SetBlock(block)
	}

	return t.remote.SetBlock(loc, block)
}

func (t *BlockRingTransport) TransferBlock(id []byte, src, dst *structs.Location) error {
	if dst.Vnode.Host == t.host {
		return errNoLocalTransfer
	}
	return t.remote.TransferBlock(id, src, dst)
}

func (t *BlockRingTransport) ReleaseBlock(loc *structs.Location, id []byte) error {
	if loc.Vnode.Host == t.host {
		return t.local.ReleaseBlock(id)
	}
	return t.remote.ReleaseBlock(loc, id)
}
