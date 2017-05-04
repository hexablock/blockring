package blockring

import (
	"github.com/hexablock/blockring/structs"
	"github.com/hexablock/blockring/utils"
)

// LocationResponse is contains a single response from a particular Location.
type LocationResponse struct {
	Location *structs.Location
	Data     interface{}
}

type RingStore struct {
	Ring     *ChordRing
	Store    *StoreTransport
	Replicas int
}

// SetBlock writes the block to the ring with the specified replicas
func (br *RingStore) SetBlock(block *structs.Block) ([]*LocationResponse, error) {
	id := block.ID()

	locs, err := br.Ring.LocateReplicatedHash(id, br.Replicas)
	if err != nil {
		return nil, err
	}

	out := make([]*LocationResponse, len(locs))
	for i, loc := range locs {
		out[i] = &LocationResponse{Location: loc}
		out[i].Data = br.Store.SetBlock(loc, block)
	}
	return out, nil
}

func (br *RingStore) GetBlock(id []byte) (*structs.Location, *structs.Block, error) {

	locs, err := br.Ring.LocateReplicatedHash(id, br.Replicas)
	if err != nil {
		return nil, nil, err
	}

	for _, loc := range locs {
		blk, er := br.Store.GetBlock(loc, id)
		if er != nil {
			err = utils.MergeErrors(err, er)
			continue
		}

		return loc, blk, nil
	}
	return nil, nil, err
}
