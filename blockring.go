package blockring

import (
	"errors"

	"github.com/hexablock/blockring/rpc"
	"github.com/hexablock/blockring/structs"
	"github.com/hexablock/blockring/utils"
	chord "github.com/ipkg/go-chord"
)

// Locator implements a location and lookup service
type Locator interface {
	LookupHash([]byte, int) (*chord.Vnode, []*chord.Vnode, error)
	LookupKey([]byte, int) ([]byte, *chord.Vnode, []*chord.Vnode, error)
	LocateReplicatedHash([]byte, int) ([]*structs.Location, error)
}

type locatorRouter struct {
	Locator
}

// RouteHash routes a hash around the ring visiting each vnode. It starts by looking up first n vnodes for the id.  If
// the first batch of n vnodes does not contain the block, then the next n vnodes from the last vnode of the
// previous batch is used until a full circle has been made around the ring.
func (rl *locatorRouter) RouteHash(hash []byte, n int, f func(*structs.Location) bool) error {
	lid := hash

	_, vns, err := rl.LookupHash(lid, n)
	if err != nil {
		return err
	}

	// Try the primary vnode
	svn := vns[0]
	pstart := int32(0)

	loc := &structs.Location{Id: hash, Vnode: svn, Priority: pstart}
	if !f(loc) {
		return nil
	}

	// Try successors in n batches
	m := map[string]bool{}
	wset := vns[1:]
	done := false
	pstart++

	for {
		// exclude vnodes we have visited.
		out := make([]*chord.Vnode, 0, n)
		for _, vn := range wset {
			// If we are back at the starting vnode we've completed a full round.
			// Set to exit after this iteration
			if utils.EqualBytes(svn.Id, vn.Id) {
				done = true
				break
			}

			// if we have not visited - track and add to list of vnodes
			if _, ok := m[vn.StringID()]; !ok {
				m[vn.StringID()] = true
				out = append(out, vn)
			}

		}

		for _, vn := range out {
			loc := &structs.Location{Id: hash, Vnode: vn, Priority: pstart}
			if !f(loc) {
				return nil
			}
			pstart++
		}

		if done {
			return nil
		}

		lid = out[len(out)-1].Id

		// update the next set of vnodes to query
		_, vn, err := rl.LookupHash(lid, n)
		if err != nil {
			return err
		}
		wset = vn
	}
}

// BlockRing is the core interface to perform operations around the ring.
type BlockRing struct {
	locator   *locatorRouter
	transport Transport

	ch               chan<- *rpc.BlockRPCData // send only channel for block transfer requests
	proxShiftEnabled bool                     // proximity shifting
}

// NewBlockRing instantiates an instance.  If the channel is not nil, proximity shifting is
// automatically enabled.
func NewBlockRing(locator Locator, transport Transport, ch chan<- *rpc.BlockRPCData) *BlockRing {
	rs := &BlockRing{
		locator:   &locatorRouter{Locator: locator},
		transport: transport,
	}
	if ch != nil {
		rs.ch = ch
		rs.proxShiftEnabled = true
	}
	return rs
}

// EnableProximityShifting enables or disables proximity shifting.  Proximity shifing can only enabled
// if the input block channel is not nil.
func (br *BlockRing) EnableProximityShifting(enable bool) {
	if enable {
		if br.ch != nil {
			br.proxShiftEnabled = true
		}
	} else {
		br.proxShiftEnabled = false
	}
}

// SetBlock writes the block to the ring with the specified replicas
func (br *BlockRing) SetBlock(block *structs.Block, opts ...RequestOptions) (*structs.Location, error) {

	o := DefaultRequestOptions()
	if len(opts) > 0 {
		o = opts[0]
	}

	id := block.ID()

	_, vns, err := br.locator.LookupHash(id, o.PeerRange)
	if err != nil {
		return nil, err
	}

	loc := &structs.Location{Id: id, Vnode: vns[0], Priority: 0}
	err = br.transport.SetBlock(loc, block)
	return loc, err
}

// GetRootBlock gets a root block with the given id
func (br *BlockRing) GetRootBlock(id []byte, opts ...RequestOptions) (*structs.Location, *structs.RootBlock, error) {
	loc, block, err := br.GetBlock(id, opts...)
	if err == nil {
		var rb structs.RootBlock
		err = rb.DecodeBlock(block)
		return loc, &rb, err
	}
	return loc, nil, err
}

// GetBlock lookups up the id hash then uses upto max successors to find the block.
func (br *BlockRing) GetBlock(id []byte, opts ...RequestOptions) (*structs.Location, *structs.Block, error) {
	o := DefaultRequestOptions()
	if len(opts) > 0 {
		o = opts[0]
	}

	var (
		blk *structs.Block
		loc *structs.Location
	)

	err := br.locator.RouteHash(id, o.PeerRange, func(l *structs.Location) bool {
		if b, err := br.transport.GetBlock(l, id); err == nil {
			blk = b
			loc = l
			return false
		}
		return true
	})

	if err == nil {
		if blk == nil {
			err = errors.New("not found")
		}

		/*if br.proxShiftEnabled {
			br.ch <- &rpc.BlockRPCData{
				Block: blk,
				Location: &structs.Location{
					Id:       id,
					Priority: 0,
					Vnode:    vn,
				},
			}
		}*/
	}

	return loc, blk, err
}
