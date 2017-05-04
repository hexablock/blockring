package structs

import (
	"fmt"
	"testing"
)

func TestRootBlock(t *testing.T) {
	ib := NewRootBlock()

	for i := uint64(0); i < 3; i++ {
		b := &Block{Type: BlockType_DATABLOCK, Data: []byte(fmt.Sprintf("1234567%d", i))}
		ib.AddBlock(i+1, b)
	}

	if ib.sz != uint64(24) {
		t.Fatal("invalid size")
	}

	blk := ib.EncodeBlock()
	if len(blk.Data) != 8+(3*32) {
		t.Fatal("invalid size", len(blk.Data))
	}

	root := NewRootBlock()
	if err := root.DecodeBlock(blk); err != nil {
		t.Fatal(err)
	}

	for k := range root.ids {
		if _, ok := ib.ids[k]; !ok {
			t.Error(k, "not found")
		}
	}

}
