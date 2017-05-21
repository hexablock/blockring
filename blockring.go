package blockring

import (
	"errors"

	"github.com/hexablock/blockring/rpc"
	"github.com/hexablock/blockring/structs"
)

// BlockRing is the core interface to perform operations around the ring.
type BlockRing struct {
	locator *locatorRouter

	blkTrans BlockTransport

	ch               chan<- *rpc.BlockRPCData // send only channel for block transfer requests
	proxShiftEnabled bool                     // proximity shifting
}

// NewBlockRing instantiates an instance.  If the channel is not nil, proximity shifting is
// automatically enabled.
func NewBlockRing(locator Locator, transport BlockTransport, ch chan<- *rpc.BlockRPCData) *BlockRing {
	rs := &BlockRing{
		locator:  &locatorRouter{Locator: locator},
		blkTrans: transport,
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
	err = br.blkTrans.SetBlock(loc, block)
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
		if b, err := br.blkTrans.GetBlock(l, id); err == nil {
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
