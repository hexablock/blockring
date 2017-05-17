package blockring

import (
	"golang.org/x/net/context"

	"github.com/hexablock/blockring/pool"
	"github.com/hexablock/blockring/rpc"
	"github.com/hexablock/blockring/store"
	"github.com/hexablock/blockring/structs"
	chord "github.com/ipkg/go-chord"
)

type LookupServiceClient struct {
	out *pool.OutConnPool
}

func NewLookupServiceClient(reapInterval, maxIdle int) *LookupServiceClient {
	return &LookupServiceClient{out: pool.NewOutConnPool(reapInterval, maxIdle)}
}

func (ls *LookupServiceClient) LookupHash(host string, hash []byte, n int) (*chord.Vnode, []*chord.Vnode, error) {
	conn, err := ls.out.Get(host)
	if err != nil {
		return nil, nil, err
	}

	defer ls.out.Return(conn)

	lreq := &rpc.LocateRequest{Key: hash, N: int32(n)}
	resp, err := conn.LocateRPC.LookupHashRPC(context.Background(), lreq)
	if err != nil {
		return nil, nil, err
	}

	return resp.Predecessor, resp.Successors, nil
}

func (ls *LookupServiceClient) LookupKey(host string, key []byte, n int) ([]byte, *chord.Vnode, []*chord.Vnode, error) {
	conn, err := ls.out.Get(host)
	if err != nil {
		return nil, nil, nil, err
	}

	defer ls.out.Return(conn)

	lreq := &rpc.LocateRequest{Key: key, N: int32(n)}
	resp, err := conn.LocateRPC.LookupKeyRPC(context.Background(), lreq)
	if err != nil {
		return nil, nil, nil, err
	}

	return resp.KeyHash, resp.Predecessor, resp.Successors, nil
}
func (ls *LookupServiceClient) LocateReplicatedKey(host string, key []byte, r int) ([]*structs.Location, error) {
	conn, err := ls.out.Get(host)
	if err != nil {
		return nil, err
	}

	defer ls.out.Return(conn)

	lreq := &rpc.LocateRequest{Key: key, N: int32(r)}
	resp, err := conn.LocateRPC.LocateReplicatedKeyRPC(context.Background(), lreq)
	if err != nil {
		return nil, err
	}
	return resp.Locations, err
}
func (ls *LookupServiceClient) LocateReplicatedHash(host string, hash []byte, r int) ([]*structs.Location, error) {
	conn, err := ls.out.Get(host)
	if err != nil {
		return nil, err
	}

	defer ls.out.Return(conn)

	lreq := &rpc.LocateRequest{Key: hash, N: int32(r)}
	resp, err := conn.LocateRPC.LocateReplicatedHashRPC(context.Background(), lreq)
	if err != nil {
		return nil, err
	}
	return resp.Locations, err
}

func (ls *LookupServiceClient) Negotiate(host string) (*rpc.NegotiateResponse, error) {
	conn, err := ls.out.Get(host)
	if err != nil {
		return nil, err
	}

	defer ls.out.Return(conn)

	req := &rpc.NegotiateRequest{}
	return conn.LocateRPC.NegotiateRPC(context.Background(), req)
}

type LookupService struct {
	ring *ChordRing
	ps   store.PeerStore
}

func NewLookupSerivce(ring *ChordRing, peerStore store.PeerStore) *LookupService {
	return &LookupService{ring: ring, ps: peerStore}
}

func (ls *LookupService) LookupKeyRPC(ctx context.Context, req *rpc.LocateRequest) (*rpc.LocateResponse, error) {
	var (
		resp = &rpc.LocateResponse{}
		err  error
	)

	resp.KeyHash, resp.Predecessor, resp.Successors, err = ls.ring.LookupKey(req.Key, int(req.N))
	return resp, err
}
func (ls *LookupService) LookupHashRPC(ctx context.Context, req *rpc.LocateRequest) (*rpc.LocateResponse, error) {
	var (
		resp = &rpc.LocateResponse{}
		err  error
	)

	resp.Predecessor, resp.Successors, err = ls.ring.LookupHash(req.Key, int(req.N))
	return resp, err
}
func (ls *LookupService) LocateReplicatedKeyRPC(ctx context.Context, req *rpc.LocateRequest) (*rpc.LocateResponse, error) {
	var (
		resp = &rpc.LocateResponse{}
		err  error
	)

	resp.Locations, err = ls.ring.LocateReplicatedKey(req.Key, int(req.N))
	return resp, err
}
func (ls *LookupService) LocateReplicatedHashRPC(ctx context.Context, req *rpc.LocateRequest) (*rpc.LocateResponse, error) {
	var (
		resp = &rpc.LocateResponse{}
		err  error
	)

	resp.Locations, err = ls.ring.LocateReplicatedHash(req.Key, int(req.N))
	return resp, err
}

func (ls *LookupService) NegotiateRPC(ctx context.Context, req *rpc.NegotiateRequest) (*rpc.NegotiateResponse, error) {
	resp := &rpc.NegotiateResponse{
		Successors: int32(ls.ring.conf.NumSuccessors),
		Vnodes:     int32(ls.ring.conf.NumVnodes),
		Peers:      ls.ps.Peers(),
	}

	return resp, nil
}
