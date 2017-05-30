package structs

import (
	"fmt"
	"testing"
)

func TestBlock(t *testing.T) {
	blk := NewDataBlock([]byte("foo"))
	b, err := blk.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	if b == nil || len(b) < 1 {
		t.Fatal("no data")
	}

	var b2 Block
	if err = b2.UnmarshalBinary(b); err != nil {
		t.Fatal(err)
	}

	if b2.Type != BlockType_DATA {
		t.Fatalf("wrong block type have=%s", b2.Type)
	}
}

func TestRootBlock(t *testing.T) {
	ib := NewRootBlock()

	for i := uint64(0); i < 3; i++ {
		b := &Block{Type: BlockType_DATA, Data: []byte(fmt.Sprintf("1234567%d", i))}
		ib.AddBlock(i+1, b)
	}

	if ib.size != uint64(24) {
		t.Fatal("invalid size")
	}

	blk := ib.EncodeBlock()
	if len(blk.Data) != fixedHeaderSize+(3*32) {
		t.Fatal("invalid size", len(blk.Data))
	}

	root := NewRootBlock()
	if err := root.DecodeBlock(blk); err != nil {
		t.Fatal(err)
	}

	for k := range root.blocks {
		if _, ok := ib.blocks[k]; !ok {
			t.Error(k, "not found")
		}
	}

	if root.Size() != ib.Size() {
		t.Fatal("size mismatch")
	}
	if root.BlockSize() != ib.BlockSize() {
		t.Fatal("block size mismatch")
	}

}
