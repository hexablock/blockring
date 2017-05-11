package blockring

import (
	"golang.org/x/net/context"

	"github.com/hexablock/blockring/pool"
	"github.com/hexablock/blockring/rpc"
	"github.com/hexablock/blockring/structs"
)

type LookupServiceClient struct {
	out *pool.OutConnPool
}

func NewLookupServiceClient(reapInterval, maxIdle int) *LookupServiceClient {
	return &LookupServiceClient{out: pool.NewOutConnPool(reapInterval, maxIdle)}
}

func (ls *LookupServiceClient) LocateBlock(host string, id []byte) ([]*structs.Location, error) {
	conn, err := ls.out.Get(host)
	if err != nil {
		return nil, err
	}

	lreq := &rpc.LocateRequest{Key: id, N: 3}
	locs, err := conn.LocateRPC.LocateReplicatedHashRPC(context.Background(), lreq)
	ls.out.Return(conn)

	return locs.Locations, err
}

type LookupService struct {
	ring *ChordRing
}

func NewLookupSerivce(ring *ChordRing) *LookupService {
	return &LookupService{ring: ring}
}

func (ls *LookupService) LocateKeyRPC(ctx context.Context, req *rpc.LocateRequest) (*rpc.LocateResponse, error) {
	resp := &rpc.LocateResponse{}

	_, pred, succ, err := ls.ring.LocateKey(req.Key, int(req.N))
	if err == nil {
		resp.Predecessor = pred
		resp.Successors = succ
	}
	return resp, err
}

func (ls *LookupService) LocateHashRPC(ctx context.Context, req *rpc.LocateRequest) (*rpc.LocateResponse, error) {
	resp := &rpc.LocateResponse{}

	pred, succ, err := ls.ring.LocateHash(req.Key, int(req.N))
	if err == nil {
		resp.Predecessor = pred
		resp.Successors = succ
	}
	return resp, err
}

func (ls *LookupService) LocateReplicatedKeyRPC(ctx context.Context, req *rpc.LocateRequest) (*rpc.LocateResponse, error) {
	resp := &rpc.LocateResponse{}

	locs, err := ls.ring.LocateReplicatedKey(req.Key, int(req.N))
	if err == nil {
		resp.Locations = locs
	}
	return resp, err
}
func (ls *LookupService) LocateReplicatedHashRPC(ctx context.Context, req *rpc.LocateRequest) (*rpc.LocateResponse, error) {
	resp := &rpc.LocateResponse{}

	locs, err := ls.ring.LocateReplicatedHash(req.Key, int(req.N))
	if err == nil {
		resp.Locations = locs
	}
	return resp, err
}
