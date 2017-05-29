package structs

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/btcsuite/fastsha256"
	"github.com/hexablock/blockring/utils"
)

// LogBlockHeader is the header for a LogBlock.
type LogBlockHeader struct {
	Source    []byte // public key of origination
	Timestamp uint64
}

// LogBlock is the log/history of a key.  It contains an ordered entry id for each operation
// applied to the key
type LogBlock struct {
	Key       []byte
	Header    *LogBlockHeader
	Root      []byte   // root hash of all entries
	Entries   [][]byte // entry hash ids
	Signature []byte
}

func NewLogBlockHeader() *LogBlockHeader {
	return &LogBlockHeader{Timestamp: uint64(time.Now().UnixNano())}
}

func (lbh LogBlockHeader) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Timestamp": lbh.Timestamp,
		"Source":    hex.EncodeToString(lbh.Source),
	})
}

func (lbh *LogBlockHeader) MarshalBinary() ([]byte, error) {
	tb := make([]byte, 8)
	binary.BigEndian.PutUint64(tb, lbh.Timestamp)
	return utils.ConcatByteSlices(tb, lbh.Source), nil
}

func (lbh *LogBlockHeader) UnmarshalBinary(b []byte) error {
	if len(b) < 8 {
		return errInvalidBlockData
	}

	lbh.Timestamp = binary.BigEndian.Uint64(b[:8])
	if len(b) > 8 {
		lbh.Source = b[8:]
	}
	return nil
}

// NewLogBlock instantiates a new empty Block.
func NewLogBlock(key []byte) *LogBlock {
	return &LogBlock{
		Key:     key,
		Header:  NewLogBlockHeader(),
		Root:    make([]byte, 32),
		Entries: make([][]byte, 0),
	}
}

// EncodeBlock encodes the LogBlock into a Block.
func (t *LogBlock) EncodeBlock() (*Block, error) {
	d, _ := t.MarshalBinary()
	return &Block{Type: BlockType_LOG, Data: d}, nil
}

// DecodeBlock decodes a block into a LogBlock
func (t *LogBlock) DecodeBlock(blk *Block) error {
	if blk.Type != BlockType_LOG {
		return ErrInvalidBlockType
	}
	return t.UnmarshalBinary(blk.Data)
}

// MarshalBinary marshals the block into a byte slice.
func (t *LogBlock) MarshalBinary() ([]byte, error) {
	// ssize
	ssize := make([]byte, 2)
	binary.BigEndian.PutUint16(ssize, uint16(len(t.Signature)))

	// header
	headerBytes, _ := t.Header.MarshalBinary()
	hsize := make([]byte, 2)
	binary.BigEndian.PutUint16(hsize, uint16(len(headerBytes)))

	// entry count
	tc := make([]byte, 4)
	binary.BigEndian.PutUint32(tc, uint32(len(t.Entries)))
	// entries
	txs := utils.ConcatByteSlices(t.Entries...)
	return utils.ConcatByteSlices(ssize, t.Signature, hsize, headerBytes, t.Root, tc, txs, t.Key), nil
}

// UnmarshalBinary unmarshals bytes into a block that were marshalled by MarshalBinary
func (t *LogBlock) UnmarshalBinary(b []byte) error {
	if len(b) < 40 {
		return errInvalidBlockData
	}
	// sig
	ssize := binary.BigEndian.Uint16(b[:2])
	t.Signature = b[2 : ssize+2]
	marker := ssize + 2
	// header
	hsize := binary.BigEndian.Uint16(b[marker : marker+2])
	marker += 2
	headerBytes := b[marker : marker+hsize]
	marker += hsize
	t.Header = &LogBlockHeader{}
	if err := t.Header.UnmarshalBinary(headerBytes); err != nil {
		return err
	}
	// root
	t.Root = b[marker : marker+32]
	marker += 32
	// entries
	tc := binary.BigEndian.Uint32(b[marker : marker+4])
	t.Entries = make([][]byte, tc)
	marker += 4
	for i := 0; i < int(tc); i++ {
		t.Entries[i] = b[marker : marker+32]
		marker += 32
	}
	t.Key = b[marker:]
	return nil
}

// LastEntry returns the last transaction id in the block.
func (t *LogBlock) LastEntry() []byte {
	l := len(t.Entries)
	if l > 0 {
		return t.Entries[l-1]
	}
	// There is none
	return nil
}

// ContainsEntry returns true if the block contains the tx hash id.
func (t *LogBlock) ContainsEntry(h []byte) bool {
	for _, th := range t.Entries {
		if utils.EqualBytes(h, th) {
			return true
		}
	}
	return false
}

// AppendEntry appends a transaction id to the block and re-calculates the merkle root if it is not found.
// If the tx id already exists it simply returns
func (t *LogBlock) AppendEntry(entry *LogEntryBlock) error {
	id := entry.ID()
	// return if we already have it.
	if t.ContainsEntry(id) {
		return nil
	}
	// check previous hash
	if len(t.Entries) > 0 {
		if !utils.EqualBytes(t.Entries[len(t.Entries)-1], entry.Header.PrevHash) {
			return ErrPrevHash
		}
	}

	// Add to block
	t.Entries = append(t.Entries, id)

	// Generate root hash
	sh := fastsha256.Sum256(utils.ConcatByteSlices(t.Entries...))
	t.Root = sh[:]
	//t.Root = merkle.GenerateTree(t.Entries).Root().Hash()
	return nil
}

// Height returns the number of tx's in the block i.e. the txblock height
func (t *LogBlock) Height() int {
	return len(t.Entries)
}

// MarshalJSON custom marshal LogBlock to be human readable.
func (t *LogBlock) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"Key":       string(t.Key),
		"Header":    t.Header,
		"Root":      hex.EncodeToString(t.Root),
		"Signature": hex.EncodeToString(t.Signature),
	}

	s := make([]string, len(t.Entries))
	for i, v := range t.Entries {
		s[i] = fmt.Sprintf("%x", v)
	}
	m["Entries"] = s

	return json.Marshal(m)
}
