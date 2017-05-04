package blockring

import (
	"golang.org/x/net/context"

	"github.com/hexablock/blockring/pool"
	"github.com/hexablock/blockring/rpc"
	"github.com/hexablock/blockring/structs"
)

type NetTransportClient struct {
	out *pool.OutConnPool
}

func NewNetTransportClient() *NetTransportClient {
	return &NetTransportClient{out: pool.NewOutConnPool(30, 180)}
}

func (s *NetTransportClient) GetBlock(loc *structs.Location, id []byte) (*structs.Block, error) {
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

func (s *NetTransportClient) SetBlock(loc *structs.Location, block *structs.Block) error {
	conn, err := s.out.Get(loc.Vnode.Host)
	if err != nil {
		return err
	}
	defer s.out.Return(conn)

	req := &rpc.BlockRPCData{Block: block}
	_, err = conn.BlockRPC.SetBlockRPC(context.Background(), req)
	return err
}

func (s *NetTransportClient) TransferBlock(loc *structs.Location, block *structs.Block) error {
	conn, err := s.out.Get(loc.Vnode.Host)
	if err != nil {
		return err
	}
	defer s.out.Return(conn)

	req := &rpc.BlockRPCData{Block: block, Location: loc}
	_, err = conn.BlockRPC.TransferBlockRPC(context.Background(), req)
	return err
}

type NetTransport struct {
	st Store
	// potential inbound blocks
	inBlocks chan *rpc.BlockRPCData
}

func NewNetTransport() *NetTransport {
	return &NetTransport{}
}

func (s *NetTransport) RegisterStore(st Store) {
	s.st = st
}

func (s *NetTransport) RegisterTakeoverQ(ch chan *rpc.BlockRPCData) {
	s.inBlocks = ch
}

func (s *NetTransport) GetBlockRPC(ctx context.Context, in *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {

	blk, err := s.st.GetBlock(in.ID)
	if err != nil {
		return nil, err
	}

	return &rpc.BlockRPCData{Block: blk}, nil
}

func (s *NetTransport) SetBlockRPC(ctx context.Context, in *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {
	err := s.st.SetBlock(in.Block)
	return &rpc.BlockRPCData{}, err
}

func (s *NetTransport) TransferBlockRPC(ctx context.Context, in *rpc.BlockRPCData) (*rpc.BlockRPCData, error) {

	s.inBlocks <- in
	return &rpc.BlockRPCData{}, nil
}
