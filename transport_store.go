package blockring

import (
	"errors"

	"golang.org/x/net/context"

	"github.com/hexablock/blockring/pool"
	"github.com/hexablock/blockring/rpc"
	"github.com/hexablock/blockring/structs"
)

var (
	errNoLocalTransfer = errors.New("local transfers not allowed")
)

// BlockStore implements a block storage interface
type BlockStore interface {
	GetBlock(id []byte) (*structs.Block, error)
	SetBlock(block *structs.Block) error
	// Marks a block to be released from the store.
	ReleaseBlock(id []byte) error
	// Iterate over all blocks
	IterBlocks(f func(block *structs.Block) error) error
	// Iteraters over all blocks in the store
	IterBlockIDs(f func([]byte) error) error
}

// BlockStoreTransport implements the transport interface for the block store.
type BlockTransport interface {
	GetBlock(loc *structs.Location, id []byte) (*structs.Block, error)
	SetBlock(loc *structs.Location, block *structs.Block) error
	TransferBlock(loc *structs.Location, id []byte) error
	ReleaseBlock(loc *structs.Location, id []byte) error
}

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

func (s *BlockNetTransportClient) TransferBlock(loc *structs.Location, id []byte) error {
	conn, err := s.out.Get(loc.Vnode.Host)
	if err != nil {
		return err
	}

	// TODO: use only block ids
	req := &rpc.BlockRPCData{Location: loc, ID: id}
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
	st BlockStore
	// potential inbound blocks
	inBlocks chan *rpc.BlockRPCData
}

func NewBlockNetTransport(bs BlockStore) *BlockNetTransport {
	return &BlockNetTransport{st: bs}
}

// Register registers a channel where incoming blocks are sent for processing.
func (s *BlockNetTransport) Register(ch chan *rpc.BlockRPCData) {
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

func (s *BlockNetTransport) TransferBlockRPC(ctx context.Context, in *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {

	s.inBlocks <- in
	return &rpc.BlockRPCData{}, nil
}

func (s *BlockNetTransport) ReleaseBlockRPC(ctx context.Context, in *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {
	err := s.st.ReleaseBlock(in.ID)
	return &rpc.BlockRPCData{}, err
}

// BlockRingTransport allows to make requests based on Location around the ring.
type BlockRingTransport struct {
	host   string
	local  BlockStore
	remote BlockTransport
}

func NewBlockRingTransport(hostname string, local BlockStore, remote BlockTransport) *BlockRingTransport {
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

func (t *BlockRingTransport) TransferBlock(loc *structs.Location, id []byte) error {
	if loc.Vnode.Host == t.host {
		return errNoLocalTransfer
	}
	return t.remote.TransferBlock(loc, id)
}

func (t *BlockRingTransport) ReleaseBlock(loc *structs.Location, id []byte) error {
	if loc.Vnode.Host == t.host {
		return t.local.ReleaseBlock(id)
	}
	return t.remote.ReleaseBlock(loc, id)
}
