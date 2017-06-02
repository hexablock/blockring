package store

import "github.com/hexablock/blockring/structs"

// LogEntryStore implements a LogEntryStore using the BlockStore interface as the backend to store
// log entries as blocks around the ring.
type LogEntryStore struct {
	bs BlockStore
}

// NewLogEntryStore instantiates a new LogEntryStore backed by the BlockStore
func NewLogEntryStore(bs BlockStore) *LogEntryStore {
	return &LogEntryStore{bs: bs}
}

// Set sets a LogEntryBlock
func (store *LogEntryStore) Set(tb *structs.LogEntryBlock) error {
	block, err := tb.EncodeBlock()
	if err == nil {
		err = store.bs.SetBlock(block)
	}
	return err
}

// Remove removes an entry block
func (store *LogEntryStore) Remove(id []byte) error {
	return store.bs.RemoveBlock(id)
}

// Get key for the tx.  This is the object encompassing the tx's for a key.
func (store *LogEntryStore) Get(id []byte) (*structs.LogEntryBlock, error) {
	block, err := store.bs.GetBlock(id)
	if err == nil {
		var le structs.LogEntryBlock
		err = le.DecodeBlock(block)
		return &le, err
	}

	return nil, err
}
