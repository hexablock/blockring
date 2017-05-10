package store

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/hexablock/blockring/structs"
)

func TestFileBlockStore(t *testing.T) {
	tf, _ := ioutil.TempDir("/tmp", "fb-test")

	defer os.RemoveAll(tf)

	fbs := NewFileBlockStore(tf)
	blk := structs.NewDataBlock([]byte("foo"))
	if err := fbs.SetBlock(blk); err != nil {
		t.Fatal(err)
	}

	gblk, err := fbs.GetBlock(blk.ID())
	if err != nil {
		t.Fatal(err)
	}

	if gblk.Type != structs.BlockType_DATABLOCK {
		t.Fatal("wrong block type")
	}

	if gblk.Size() != 3 {
		t.Fatal("wrong size")
	}
	c := 0
	fbs.IterBlocks(func(block *structs.Block) error {
		c++
		return nil
	})

	if c != 1 {
		t.Fatal("wrong block count")
	}
}
