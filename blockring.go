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

// BlockRing is the core interface to perform operations around the ring.
type BlockRing struct {
	locator   Locator
	transport Transport

	ch               chan<- *rpc.BlockRPCData // send only channel for block transfer requests
	proxShiftEnabled bool                     // proximity shifting
}

// NewBlockRing instantiates an instance.  If the channel is not nil, proximity shifting is
// automatically enabled.
func NewBlockRing(locator Locator, transport Transport, ch chan<- *rpc.BlockRPCData) *BlockRing {
	rs := &BlockRing{
		locator:   locator,
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

	return br.route(id, o.PeerRange)

}

// getBlock gets the first available block from the given vnodes. Any found block is submitted to shift proximity
// if enabled.
func (br *BlockRing) getBlock(id []byte, vns []*chord.Vnode, pstart int32) (*structs.Location, *structs.Block, error) {
	var err error

	for i, vn := range vns {

		loc := &structs.Location{Id: id, Vnode: vn, Priority: int32(i) + pstart}
		blk, er := br.transport.GetBlock(loc, id)
		if er != nil {
			//log.Printf("ERR action=GetBlock block=%x distance=%d msg='%v'", id, i, er)
			err = utils.MergeErrors(err, er)
			continue
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

		return loc, blk, nil
	}
	return nil, nil, err
}

// route routes a getblock request around the ring. It starts by looking up first n vnodes for the id.  If
// the first batch of n vnodes does not contain the block, then the next n vnodes from the last vnode of the
// previous batch is used until a full circle has been made around the ring.
func (br *BlockRing) route(id []byte, n int) (*structs.Location, *structs.Block, error) {
	lid := id

	_, vns, err := br.locator.LookupHash(lid, n)
	if err != nil {
		return nil, nil, err
	}

	// Try the primary vnode
	svn := vns[0]
	loc := &structs.Location{Id: id, Vnode: svn, Priority: 0}
	blk, err := br.transport.GetBlock(loc, id)
	if err == nil {
		return loc, blk, nil
	}

	// Try successors in n batches
	m := map[string]bool{}
	wset := vns[1:]
	bail := false
	pstart := 1

	for {
		// exclude vnodes we have visited.
		out := make([]*chord.Vnode, 0, n)
		for _, vn := range wset {
			// if we have not visited - track and add to list of vnodes
			if _, ok := m[vn.StringID()]; !ok {
				m[vn.StringID()] = true
				out = append(out, vn)
			}
			// If we are back at the starting vnode we've completed a full round.
			// Set to exit after this iteration
			if utils.EqualBytes(svn.Id, vn.Id) {
				bail = true
				break
			}

		}

		loc, blk, err := br.getBlock(id, out, int32(pstart))
		if err == nil {
			return loc, blk, nil
		}

		if bail {
			return nil, nil, errors.New("not found")
		}

		lid = out[len(out)-1].Id
		pstart += len(out)
		// update the next set of vnodes to query
		_, vn, err := br.locator.LookupHash(lid, n)
		if err != nil {
			return nil, nil, err
		}
		wset = vn
	}
}
