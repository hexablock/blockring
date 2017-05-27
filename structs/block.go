package structs

import (
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/btcsuite/fastsha256"
)

var (
	// ErrInvalidBlockType is used when the block has an unrecognized type
	ErrInvalidBlockType = errors.New("invalid block type")
	errInvalidBlockData = errors.New("invalid block data")
)

// NewDataBlock instantiates a new data block with the given data
func NewDataBlock(data []byte) *Block {
	return &Block{Type: BlockType_DATA, Data: data}
}

// MarshalBinary marshals the block into a byte slice.
func (blk *Block) MarshalBinary() ([]byte, error) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(blk.Type))
	return append([]byte{b[3]}, blk.Data...), nil
}

// UnmarshalBinary unmarshals bytes into a block.  The first byte is the type with the remainder
// being the data.
func (blk *Block) UnmarshalBinary(b []byte) error {
	if len(b) < 2 {
		return ErrInvalidBlockType
	}

	t := binary.BigEndian.Uint32([]byte{0, 0, 0, b[0]})
	blk.Type = BlockType(t)
	blk.Data = b[1:]
	return nil
}

// MarshalJSON custom json marshaller
func (blk *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Type": blk.Type,
		"Data": blk.Data,
		"Size": len(blk.Data),
	})
}

// ID returns the hash id of the block
func (blk *Block) ID() []byte {
	d, _ := blk.MarshalBinary()
	sh := fastsha256.Sum256(d)
	return sh[:]
}

// Size returns the size of the data if a DATABLOCK and the cummulative size of all blocks if it is
// a RootBlock
func (blk *Block) Size() uint64 {
	switch blk.Type {
	case BlockType_ROOT:
		if len(blk.Data) < 8 {
			return 0
		}
		return binary.BigEndian.Uint64(blk.Data[:8])
	}
	return uint64(len(blk.Data))
}
