package store

import (
	"errors"
	"sync"

	"github.com/hexablock/blockring/structs"
)

var (
	errNotFound = errors.New("not found")
)

// NewMemBlockStore instantiates a new in-memory Block store.
func NewMemBlockStore() *MemBlockStore {
	return &MemBlockStore{m: make(map[string]*structs.Block)}
}

// MemBlockStore is an in-memory block store.
type MemBlockStore struct {
	mu sync.RWMutex
	m  map[string]*structs.Block
}

// GetBlock returns a block with the given id if it exists
func (mem *MemBlockStore) GetBlock(id []byte) (*structs.Block, error) {

	mem.mu.RLock()
	if v, ok := mem.m[string(id)]; ok {
		mem.mu.RUnlock()
		return v, nil
	}
	mem.mu.RUnlock()
	return nil, errNotFound
}

// IterBlocks iterates over blocks in theh store.  If an error is returned by the callback
// iteration is immediately terminated returning the error.
func (mem *MemBlockStore) IterBlocks(f func(block *structs.Block) error) error {
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

// SetBlock writes the given block to the store returning an error on failure
func (mem *MemBlockStore) SetBlock(blk *structs.Block) error {
	mem.mu.Lock()
	mem.m[string(blk.ID())] = blk
	mem.mu.Unlock()

	return nil
}
