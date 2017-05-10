package structs

import (
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/btcsuite/fastsha256"
	"github.com/hexablock/blockring/utils"
)

var (
	ErrInvalidBlockType = errors.New("invalid block type")
)

func NewDataBlock(data []byte) *Block {
	return &Block{Type: BlockType_DATABLOCK, Data: data}
}

func (blk *Block) MarshalBinary() ([]byte, error) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(blk.Type))
	return append([]byte{b[3]}, blk.Data...), nil
}
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

// ID return the hash of the block
func (blk *Block) ID() []byte {
	bs := make([]byte, 2)
	binary.BigEndian.PutUint16(bs, uint16(blk.Type))

	d := utils.ConcatByteSlices(bs, blk.Data)
	//d := append(bs, blk.Data...)

	sh := fastsha256.Sum256(d)
	return sh[:]
}

// Size returns the size of the data if a DATABLOCK and the cummulative size of all blocks if it is
// an INDEXBLOCK
func (blk *Block) Size() uint64 {
	switch blk.Type {
	case BlockType_ROOTBLOCK:
		if len(blk.Data) < 8 {
			return 0
		}
		return binary.BigEndian.Uint64(blk.Data[:8])
	}
	return uint64(len(blk.Data))
}
