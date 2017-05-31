package blockring

import (
	"sync"
	"sync/atomic"

	"github.com/hexablock/blockring/rpc"
	"github.com/hexablock/blockring/structs"
	"github.com/hexablock/blockring/utils"
)

// LogTransport implements a transport for the distributed log
type LogTransport interface {
	GetEntry(loc *structs.Location, key []byte, opts structs.RequestOptions) (*structs.LogEntryBlock, *structs.Location, error)
	NewEntry(loc *structs.Location, key []byte, opts structs.RequestOptions) (*structs.LogEntryBlock, *structs.Location, error)
	ProposeEntry(loc *structs.Location, tx *structs.LogEntryBlock, opts structs.RequestOptions) (*structs.Location, error)
	CommitEntry(loc *structs.Location, tx *structs.LogEntryBlock, opts structs.RequestOptions) (*structs.Location, error)
	GetLogBlock(loc *structs.Location, key []byte, opts structs.RequestOptions) (*structs.LogBlock, *structs.Location, error)
	TransferLogBlock(key []byte, src, dst *structs.Location) error
}

// BlockTransport implements the transport interface for the block store.
type BlockTransport interface {
	GetBlock(loc *structs.Location, id []byte) (*structs.Block, error)
	SetBlock(loc *structs.Location, block *structs.Block) error
	TransferBlock(id []byte, src, dst *structs.Location) error
	ReleaseBlock(loc *structs.Location, id []byte) error
}

// BlockRing is the core interface to perform operations around the ring.
type BlockRing struct {
	locator *locatorRouter // This could be client or server side locator

	blkTrans BlockTransport
	logTrans LogTransport

	ch               chan<- *rpc.RelocateRPCData // Send only channel for block transfer requests
	proxShiftEnabled bool                        // Proximity shifting
}

// NewBlockRing instantiates an instance.  If the channel is not nil, proximity shifting is
// automatically enabled.
func NewBlockRing(locator Locator, blkTrans BlockTransport, logTrans LogTransport, ch chan<- *rpc.RelocateRPCData) *BlockRing {
	rs := &BlockRing{
		locator:  &locatorRouter{Locator: locator},
		blkTrans: blkTrans,
		logTrans: logTrans,
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
func (br *BlockRing) SetBlock(block *structs.Block, opts ...structs.RequestOptions) (*structs.Location, error) {

	o := structs.DefaultRequestOptions()
	if len(opts) > 0 {
		o = &opts[0]
	}

	id := block.ID()

	_, vns, err := br.locator.LookupHash(id, int(o.PeerSetSize))
	if err != nil {
		return nil, err
	}

	loc := &structs.Location{Id: id, Vnode: vns[0], Priority: 0}
	err = br.blkTrans.SetBlock(loc, block)
	return loc, err
}

// GetBlock lookups up the id hash then uses upto max successors to find the block.
func (br *BlockRing) GetBlock(id []byte, opts ...structs.RequestOptions) (*structs.Location, *structs.Block, error) {
	o := structs.DefaultRequestOptions()
	if len(opts) > 0 {
		o = &opts[0]
	}

	var (
		blk *structs.Block
		loc *structs.Location
	)

	err := br.locator.RouteHash(id, int(o.PeerSetSize), func(l *structs.Location) bool {

		if b, err := br.blkTrans.GetBlock(l, id); err == nil {
			blk = b
			loc = l
			return false
		}
		return true
	})

	if err == nil {
		if blk == nil {
			err = utils.ErrNotFound
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

// GetRootBlock gets a root block with the given id
func (br *BlockRing) GetRootBlock(id []byte, opts ...structs.RequestOptions) (*structs.Location, *structs.RootBlock, error) {
	loc, block, err := br.GetBlock(id, opts...)
	if err == nil {
		var rb structs.RootBlock
		err = rb.DecodeBlock(block)
		return loc, &rb, err
	}
	return loc, nil, err
}

// GetLogBlock gets the LogBlock by routing the key until it is found.
func (br *BlockRing) GetLogBlock(key []byte, opts structs.RequestOptions) (*structs.Location, *structs.LogBlock, error) {

	var (
		blk *structs.LogBlock
		loc *structs.Location
	)

	err := br.locator.RouteKey(key, int(opts.PeerSetSize), func(l *structs.Location) bool {
		if b, _, err := br.logTrans.GetLogBlock(l, key, opts); err == nil {
			blk = b
			loc = l
			return false
		}
		return true
	})

	if err == nil {
		if blk == nil {
			err = utils.ErrNotFound
		}
		//if br.proxShiftEnabled {}
	}

	return loc, blk, err

}

// GetBlockFrom gets a Block from the specified Location
func (br *BlockRing) GetBlockFrom(id []byte, loc *structs.Location) (*structs.Block, error) {
	return br.blkTrans.GetBlock(loc, id)
}

// GetEntry gets a LogEntryBlock from the ring.
func (br *BlockRing) GetEntry(id []byte, opts structs.RequestOptions) (*structs.Location, *structs.LogEntryBlock, error) {

	var (
		blk *structs.LogEntryBlock
		loc *structs.Location
	)

	err := br.locator.RouteHash(id, int(opts.PeerSetSize), func(l *structs.Location) bool {
		if b, _, err := br.logTrans.GetEntry(l, id, opts); err == nil {
			blk = b
			loc = l
			return false
		}
		return true
	})

	if err == nil {
		if blk == nil {
			err = utils.ErrNotFound
		}
		//if br.proxShiftEnabled {}
	}

	return loc, blk, err
}

// NewEntry gets a new entry from the log.
func (br *BlockRing) NewEntry(key []byte, opts structs.RequestOptions) (*structs.LogEntryBlock, *structs.Location, error) {

	locs, err := br.locator.LocateReplicatedKey(key, int(opts.PeerSetSize))
	if err != nil {
		return nil, nil, err
	}

	// Return the first good entry from a location
	var l *structs.Location
	for _, loc := range locs {
		var blk *structs.LogEntryBlock
		if blk, _, err = br.logTrans.NewEntry(loc, key, opts); err == nil {
			return blk, loc, nil
		}
		l = loc
	}
	// Return error and associated location of err
	return nil, l, err
}

// ProposeEntry proposes a transaction to the network.
func (br *BlockRing) ProposeEntry(tx *structs.LogEntryBlock, opts structs.RequestOptions) (*structs.Location, error) {

	locs, err := br.locator.LocateReplicatedKey(tx.Key, int(opts.PeerSetSize))
	if err != nil {
		return nil, err
	}

	var (
		wg    sync.WaitGroup
		errCh = make(chan error, len(locs))
		done  = make(chan struct{})
		bail  int32
		meta  *structs.Location
	)

	wg.Add(len(locs))

	if opts.Source != nil && len(opts.Source) > 0 {
		// Broadcast to all vnodes skipping the source.
		for _, l := range locs {
			// 1 go-routine per location
			go func(loc *structs.Location) {

				if atomic.LoadInt32(&bail) == 0 {

					if !utils.EqualBytes(loc.Id, opts.Source) {
						o := structs.RequestOptions{
							Destination: loc.Id,
							Source:      opts.Source,
							PeerSetSize: opts.PeerSetSize,
							PeerSetKey:  loc.Id,
						}

						if _, er := br.logTrans.ProposeEntry(loc, tx, o); er != nil {
							errCh <- er
						}
					}

				}
				wg.Done()

			}(l)

		}

	} else {
		// Broadcast to all vnodes
		for _, l := range locs {

			go func(loc *structs.Location) {

				if atomic.LoadInt32(&bail) == 0 {
					o := structs.RequestOptions{
						Destination: loc.Id,
						Source:      loc.Id,
						PeerSetSize: opts.PeerSetSize,
						PeerSetKey:  loc.Id,
					}

					if _, er := br.logTrans.ProposeEntry(loc, tx, o); er != nil {
						errCh <- er
					}
				}

				wg.Done()

			}(l)

		}

	}

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case err = <-errCh:
		atomic.StoreInt32(&bail, 1)
	}

	return meta, err
}

// CommitEntry tries to commit an entry
func (br *BlockRing) CommitEntry(tx *structs.LogEntryBlock, opts structs.RequestOptions) (*structs.Location, error) {
	locs, err := br.locator.LocateReplicatedKey(tx.Key, int(opts.PeerSetSize))
	if err != nil {
		return nil, err
	}

	var (
		wg    sync.WaitGroup
		errCh = make(chan error, len(locs))
		done  = make(chan struct{})
		bail  int32
		meta  *structs.Location
	)

	wg.Add(len(locs))

	if opts.Source != nil && len(opts.Source) > 0 {
		// Broadcast to all vnodes skipping the source.
		for _, l := range locs {

			go func(loc *structs.Location) {
				if atomic.LoadInt32(&bail) == 0 {

					if !utils.EqualBytes(loc.Id, opts.Source) {
						o := structs.RequestOptions{
							Destination: loc.Id,
							PeerSetKey:  loc.Id,
							Source:      opts.Source,
							PeerSetSize: opts.PeerSetSize,
						}
						if _, er := br.logTrans.CommitEntry(loc, tx, o); er != nil {
							errCh <- er
						}
					}
				}
				wg.Done()
			}(l)

		}

	} else {
		// Broadcast to all vnodes
		for _, l := range locs {

			go func(loc *structs.Location) {
				if atomic.LoadInt32(&bail) == 0 {
					o := structs.RequestOptions{
						PeerSetSize: opts.PeerSetSize,
						Source:      loc.Id,
						Destination: loc.Id,
						PeerSetKey:  loc.Id,
					}

					if _, er := br.logTrans.CommitEntry(loc, tx, o); er != nil {
						errCh <- er
					}

				}
				wg.Done()
			}(l)
		}

	}

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case err = <-errCh:
		atomic.StoreInt32(&bail, 1)
	}

	return meta, err
}
