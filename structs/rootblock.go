package structs

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"os"
	"sync"

	"github.com/hexablock/blockring/utils"
)

const (
	// DefaultBlockSize is the default block size used when none is supplied.
	DefaultBlockSize = 512 * 1024
	// mode=4bytes, size=8bytes, block-size=4bytes
	fixedHeaderSize = 16
)

// RootBlock contains the index information for a dataset.
type RootBlock struct {
	mu        sync.RWMutex
	size      uint64            // Sum of data size for all ref'd blocks.
	blockSize uint32            // Size of each block. Last block will be less than this.
	mode      os.FileMode       // File mode (uint32)
	blocks    map[uint64][]byte // Block ids that belong to this block.
}

// NewRootBlock instantiates a new RootBlock with defaults.
func NewRootBlock() *RootBlock {
	return &RootBlock{
		blocks:    make(map[uint64][]byte),
		blockSize: DefaultBlockSize,
	}
}

// SetMode sets the mode on the RootBlock
func (idx *RootBlock) SetMode(mode os.FileMode) {
	idx.mode = mode
}

// Mode returns the *nix mode of the block
func (idx *RootBlock) Mode() os.FileMode {
	return idx.mode
}

// BlockSize returns the block size of the data
func (idx *RootBlock) BlockSize() uint32 {
	return idx.blockSize
}

// SetBlockSize sets the block size
func (idx *RootBlock) SetBlockSize(bs uint32) {
	idx.blockSize = bs
}

// Len returns the number of blocks in the root block
func (idx *RootBlock) Len() int {
	return len(idx.blocks)
}

// MarshalJSON is a custom json marshaller to handle hashes
func (idx *RootBlock) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"Type":      BlockType_ROOT,
		"Size":      idx.size,
		"BlockSize": idx.blockSize,
		"Mode":      idx.mode,
	}

	ids := make([]string, len(idx.blocks))
	idx.Iter(func(index uint64, id []byte) error {
		ids[index] = hex.EncodeToString(id)
		return nil
	})

	m["Blocks"] = ids
	return json.Marshal(m)
}

// Size returns the total size of the data
func (idx *RootBlock) Size() uint64 {
	return idx.size
}

// AddBlock adds a block to the RootBlock at the given index.
func (idx *RootBlock) AddBlock(index uint64, blk *Block) {
	id := blk.ID()

	idx.mu.Lock()
	idx.size += uint64(len(blk.Data))
	idx.blocks[index] = id
	idx.mu.Unlock()
}

// Iter iterates over each block id in order.
func (idx *RootBlock) Iter(f func(index uint64, id []byte) error) error {
	// sort by index
	keys := make([][]byte, len(idx.blocks))
	for i := range idx.blocks {
		keys[i-1] = idx.blocks[i]
	}

	for i, k := range keys {
		if err := f(uint64(i), k); err != nil {
			return err
		}
	}

	return nil
}

// ID returns the hash id of the block
func (idx *RootBlock) ID() []byte {
	return idx.EncodeBlock().ID()
}

// EncodeBlock encodes the RootBlock into a Block
func (idx *RootBlock) EncodeBlock() *Block {
	a := make([][]byte, len(idx.blocks))
	for i := range idx.blocks {
		a[i-1] = idx.blocks[i]
	}

	bids := utils.ConcatByteSlices(a...)

	blk := &Block{Type: BlockType_ROOT}
	// mode
	mb := make([]byte, 4)
	binary.BigEndian.PutUint32(mb, uint32(idx.mode))
	// size
	sb := make([]byte, 8)
	binary.BigEndian.PutUint64(sb, idx.size)
	// block size
	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, idx.blockSize)

	blk.Data = utils.ConcatByteSlices(mb, sb, bs, bids)
	return blk
}

// DecodeBlock decodes block data into a RootBlock
func (idx *RootBlock) DecodeBlock(block *Block) error {
	if block.Type != BlockType_ROOT {
		return ErrInvalidBlockType
	}
	if len(block.Data) < fixedHeaderSize {
		return errInvalidBlockData
	}

	idx.mode = os.FileMode(binary.BigEndian.Uint32(block.Data[:4]))
	idx.size = binary.BigEndian.Uint64(block.Data[4:12])
	idx.blockSize = binary.BigEndian.Uint32(block.Data[12:fixedHeaderSize])
	if (len(block.Data[fixedHeaderSize:]) % 32) != 0 {
		return errInvalidBlockData
	}

	idx.blocks = make(map[uint64][]byte)
	c := uint64(1)
	l := uint64(len(block.Data))
	for i := uint64(fixedHeaderSize); i < l; i += 32 {
		idx.blocks[c] = block.Data[i : i+32]
		c++
	}

	return nil
}
