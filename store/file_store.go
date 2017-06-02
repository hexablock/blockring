package store

import (
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hexablock/log"

	"github.com/hexablock/blockring/structs"
	"github.com/hexablock/blockring/utils"
)

// FileBlockStore implements a file based block store.  All writes occur on the in-memory buffer
// which is then flushed at a later interval.
type FileBlockStore struct {
	datadir        string
	defaultSetPerm os.FileMode

	mu            sync.RWMutex
	buf           map[string]*structs.Block
	flushInterval time.Duration
}

// NewFileBlockStore instantiates a new FileBlockStore setting the defaults permissions, flush interval
// provided data directory.
func NewFileBlockStore(datadir string) *FileBlockStore {
	fbs := &FileBlockStore{
		datadir:        datadir,
		defaultSetPerm: 0444,
		flushInterval:  20 * time.Second,
		buf:            make(map[string]*structs.Block),
	}
	go fbs.flushBlocks()
	return fbs
}

// RemoveBlock removes a Block from the in-mem buffer as well as stable store.
func (st *FileBlockStore) RemoveBlock(id []byte) error {
	p := st.abspath(hex.EncodeToString(id))
	err := os.Remove(p)

	st.mu.Lock()
	if _, ok := st.buf[string(id)]; ok {
		delete(st.buf, string(id))
	}
	st.mu.Unlock()
	return err
}

// GetBlock returns a block with the given id if it exists
func (st *FileBlockStore) GetBlock(id []byte) (*structs.Block, error) {
	st.mu.RLock()
	if v, ok := st.buf[string(id)]; ok {
		st.mu.RUnlock()
		return v, nil
	}
	st.mu.RUnlock()

	ap := st.abspath(hex.EncodeToString(id))
	return st.readBlockFromFile(ap)
}

// IterIDs iterates over all block ids
func (st *FileBlockStore) IterIDs(f func(id []byte) error) error {
	// Traverse in-mem blocks
	i := 0
	st.mu.RLock()
	inMem := make([][]byte, len(st.buf))
	for k := range st.buf {
		inMem[i] = []byte(k)
		i++
	}
	st.mu.RUnlock()

	for _, v := range inMem {
		if err := f(v); err != nil {
			return err
		}
	}

	// Traverse all block files
	files, err := ioutil.ReadDir(st.datadir)
	if err != nil {
		return err
	}

	for _, fl := range files {
		if fl.IsDir() {
			continue
		}

		id, er := hex.DecodeString(fl.Name())
		if er != nil {
			continue
		}

		if er := f(id); er != nil {
			err = er
			break
		}
	}

	return err
}

// Iter iterates over blocks in theh store.  If an error is returned by the callback
// iteration is immediately terminated returning the error.
func (st *FileBlockStore) Iter(f func(block *structs.Block) error) error {
	// read in-mem blocks
	st.mu.RLock()
	for _, v := range st.buf {
		if err := f(v); err != nil {
			return err
		}
	}
	st.mu.RUnlock()

	files, err := ioutil.ReadDir(st.datadir)
	if err != nil {
		return err
	}

	for _, fl := range files {
		if fl.IsDir() {
			continue
		}

		if _, er := hex.DecodeString(fl.Name()); er != nil {
			continue
		}

		fp := st.abspath(fl.Name())
		blk, er := st.readBlockFromFile(fp)
		if er != nil {
			err = er
			break
		}
		if er := f(blk); er != nil {
			err = er
			break
		}
	}

	return err
}

// SetBlock writes the given block to memory which later gets flushed to disk
func (st *FileBlockStore) SetBlock(blk *structs.Block) error {
	id := string(blk.ID())
	st.mu.Lock()
	st.buf[id] = blk
	st.mu.Unlock()

	return nil
}

// ReleaseBlock marks a block to be released (eventually removed) from the store.
func (st *FileBlockStore) ReleaseBlock(id []byte) error {
	sid := hex.EncodeToString(id)
	ap := st.abspath(sid)

	if _, err := os.Stat(ap); err != nil {
		return utils.ErrNotFound
	}

	//log.Printf("Releasing block/%x path='%s'", id, ap)

	return errors.New("TBI")
}

func (st *FileBlockStore) abspath(p string) string {
	return filepath.Join(st.datadir, p)
}

func (st *FileBlockStore) flushBlocks() {
	for {
		time.Sleep(st.flushInterval)
		// write all in-mem blocks to disk removing from in-mem on success
		st.mu.Lock()
		for k, v := range st.buf {
			if err := st.writeBlockToFile(v); err != nil {
				log.Println("[ERROR]", err)
				continue
			}
			delete(st.buf, k)
		}
		st.mu.Unlock()
	}
}

func (st *FileBlockStore) writeBlockToFile(blk *structs.Block) error {
	bid := blk.ID()
	fp := st.abspath(hex.EncodeToString(bid))

	_, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			data, _ := blk.MarshalBinary()
			return ioutil.WriteFile(fp, data, st.defaultSetPerm)
		}
	}

	return err
}

func (st *FileBlockStore) readBlockFromFile(fp string) (*structs.Block, error) {
	b, err := ioutil.ReadFile(fp)
	if err == nil {
		var blk structs.Block
		err = blk.UnmarshalBinary(b)
		return &blk, err
	}
	return nil, err
}
