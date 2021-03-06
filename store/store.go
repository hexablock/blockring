package store

import (
	"errors"
	"sync"

	"github.com/hexablock/blockring/structs"
	"github.com/hexablock/blockring/utils"
)

// BlockStore implements a block storage interface
type BlockStore interface {
	GetBlock(id []byte) (*structs.Block, error)
	RemoveBlock(id []byte) error
	SetBlock(block *structs.Block) error
	// Marks a block to be released from the store.
	ReleaseBlock(id []byte) error
	// Iterate over all blocks in the store
	Iter(f func(block *structs.Block) error) error
	// Iteraters over all block id's in the store
	IterIDs(f func(id []byte) error) error
}

// MemBlockStore is an in-memory block store.
type MemBlockStore struct {
	mu sync.RWMutex
	m  map[string]*structs.Block
}

// NewMemBlockStore instantiates a new in-memory Block store.
func NewMemBlockStore() *MemBlockStore {
	return &MemBlockStore{m: make(map[string]*structs.Block)}
}

// ReleaseBlock marks a block to be released from the store.
func (mem *MemBlockStore) ReleaseBlock(id []byte) error {

	mem.mu.RLock()
	if _, ok := mem.m[string(id)]; ok {
		mem.mu.RUnlock()
		return errors.New("TBI")
	}
	mem.mu.RUnlock()

	return utils.ErrNotFound
}

// GetBlock returns a block with the given id if it exists
func (mem *MemBlockStore) GetBlock(id []byte) (*structs.Block, error) {

	mem.mu.RLock()
	if v, ok := mem.m[string(id)]; ok {
		mem.mu.RUnlock()
		return v, nil
	}
	mem.mu.RUnlock()
	return nil, utils.ErrNotFound
}

// Iter iterates over blocks in theh store.  If an error is returned by the callback
// iteration is immediately terminated returning the error.
func (mem *MemBlockStore) Iter(f func(block *structs.Block) error) error {
	mem.mu.RLock()
	for _, b := range mem.m {
		if err := f(b); err != nil {
			mem.mu.RUnlock()
			return err
		}
	}
	mem.mu.RUnlock()

	return nil
}

func (mem *MemBlockStore) IterIDs(f func([]byte) error) error {
	mem.mu.RLock()
	for istr := range mem.m {

		if err := f([]byte(istr)); err != nil {
			mem.mu.RUnlock()
			return err
		}
	}
	mem.mu.RUnlock()

	return nil
}

// SetBlock writes the given block to the store returning an error on failure
func (mem *MemBlockStore) SetBlock(blk *structs.Block) error {
	mem.mu.Lock()
	mem.m[string(blk.ID())] = blk
	mem.mu.Unlock()

	return nil
}
