package store

import "github.com/hexablock/blockring/structs"

type LogEntryStore struct {
	bs BlockStore
}

func (les *LogEntryStore) Get(id []byte) (*structs.LogEntryBlock, error) {
	// blk, err := les.bs.GetBlock(id)
	// if err == nil {
	//
	// }
	// return nil, err
	return nil, nil
}

func (les *LogEntryStore) Set(tx *structs.LogEntryBlock) error {
	return nil
}
