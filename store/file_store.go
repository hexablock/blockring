package store

import (
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hexablock/blockring/structs"
)

type FileBlockStore struct {
	datadir        string
	defaultSetPerm os.FileMode
}

func NewFileBlockStore(datadir string) *FileBlockStore {
	return &FileBlockStore{datadir: datadir, defaultSetPerm: 0644}
}

func (st *FileBlockStore) abspath(p string) string {
	return filepath.Join(st.datadir, p)
}

// GetBlock returns a block with the given id if it exists
func (st *FileBlockStore) GetBlock(id []byte) (*structs.Block, error) {
	ap := st.abspath(hex.EncodeToString(id))
	return st.readBlockFromFile(ap)
}

// IterBlocks iterates over blocks in theh store.  If an error is returned by the callback
// iteration is immediately terminated returning the error.
func (st *FileBlockStore) IterBlocks(f func(block *structs.Block) error) error {

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

// SetBlock writes the given block to the store returning an error on failure
func (st *FileBlockStore) SetBlock(blk *structs.Block) error {
	return st.writeBlockToFile(blk)
}

func (st *FileBlockStore) writeBlockToFile(blk *structs.Block) error {
	bid := blk.ID()
	fp := st.abspath(hex.EncodeToString(bid))
	// TODO: don't write block if it exists
	data, _ := blk.MarshalBinary()
	return ioutil.WriteFile(fp, data, st.defaultSetPerm)
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
