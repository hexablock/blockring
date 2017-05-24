package structs

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"

	"github.com/hexablock/blockring/utils"
)

const DefaultBlockSize = 128 * 1024

var (
	errInvalidBlockData = errors.New("invalid block data")
)

type RootBlock struct {
	mu  sync.RWMutex
	sz  uint64
	bs  uint32 // block size
	ids map[uint64][]byte
}

func NewRootBlock() *RootBlock {
	return &RootBlock{ids: make(map[uint64][]byte), bs: DefaultBlockSize}
}

func (idx *RootBlock) BlockSize() uint32 {
	return idx.bs
}
func (idx *RootBlock) SetBlockSize(bs uint32) {
	idx.bs = bs
}

// Len returns the number of blocks in the root block
func (idx *RootBlock) Len() int {
	return len(idx.ids)
}

func (idx *RootBlock) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"Type":      BlockType_ROOT,
		"Size":      idx.sz,
		"BlockSize": idx.bs,
	}

	ids := make([]string, len(idx.ids))
	idx.Iter(func(index uint64, id []byte) error {
		ids[index] = hex.EncodeToString(id)
		return nil
	})

	m["Blocks"] = ids
	return json.Marshal(m)
}

func (idx *RootBlock) Size() uint64 {
	return idx.sz
}

// AddBlock adds a block to the RootBlock at the given index.
func (idx *RootBlock) AddBlock(index uint64, blk *Block) {
	id := blk.ID()

	idx.mu.Lock()
	idx.sz += uint64(len(blk.Data))
	idx.ids[index] = id
	idx.mu.Unlock()
}

// Iter iterates over each block id in order.
func (idx *RootBlock) Iter(f func(index uint64, id []byte) error) error {
	// sort by index
	keys := make([][]byte, len(idx.ids))
	for i := range idx.ids {
		keys[i-1] = idx.ids[i]
	}

	for i, k := range keys {
		if err := f(uint64(i), k); err != nil {
			return err
		}
	}

	return nil
}

func (idx *RootBlock) ID() []byte {
	return idx.EncodeBlock().ID()
}

// EncodeBlock encodes the RootBlock into a Block
func (idx *RootBlock) EncodeBlock() *Block {
	a := make([][]byte, len(idx.ids))
	for i := range idx.ids {
		a[i-1] = idx.ids[i]
	}

	bids := utils.ConcatByteSlices(a...)

	blk := &Block{Type: BlockType_ROOT}
	// size
	sb := make([]byte, 8)
	binary.BigEndian.PutUint64(sb, idx.sz)
	// block size
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, idx.bs)

	blk.Data = utils.ConcatByteSlices(sb, bs, bids)
	return blk
}

// DecodeBlock decodes block data into a RootBlock
func (idx *RootBlock) DecodeBlock(block *Block) error {
	if block.Type != BlockType_ROOT {
		return ErrInvalidBlockType
	}
	if len(block.Data) < 12 {
		return errInvalidBlockData
	}

	idx.sz = binary.BigEndian.Uint64(block.Data[:8])
	idx.bs = binary.BigEndian.Uint32(block.Data[8:12])
	if (len(block.Data[12:]) % 32) != 0 {
		return errInvalidBlockData
	}

	idx.ids = make(map[uint64][]byte)
	c := uint64(1)
	l := uint64(len(block.Data))
	for i := uint64(12); i < l; i += 32 {
		idx.ids[c] = block.Data[i : i+32]
		c++
	}

	return nil
}

/*

// AppendBlock adds the given block to the IndexBlock
func (ib *IndexBlock) AppendBlock(blk *Block) error {
	if blk.Type != BlockType_DATABLOCK {
		return errInvalidBlockType
	}
	id := blk.ID()
	ib.AppendBlockID(id, blk.Size())
	return nil
}

func (ib *IndexBlock) SetIDs(ids ...[]byte) {
	d := utils.ConcatByteSlices(ids...)
	ib.Data = append(ib.Data[:8], d...)
}
*/
