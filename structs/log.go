package structs

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/btcsuite/fastsha256"
	"github.com/hexablock/blockring/utils"
)

var ErrPrevHash = errors.New("prev hash mismatch")

type LogEntryHeader struct {
	PrevHash  []byte
	Source    []byte
	Timestamp uint64
	Height    uint64
}

func (header LogEntryHeader) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"PrevHash":  hex.EncodeToString(header.PrevHash),
		"Source":    hex.EncodeToString(header.Source),
		"Height":    header.Height,
		"Timestamp": header.Timestamp,
	})
}

type LogEntryBlock struct {
	Key       []byte
	Header    *LogEntryHeader
	Data      []byte
	Signature []byte
}

type LogBlock struct {
	Key  []byte
	Root []byte
	TXs  [][]byte
}

var errInvalidData = errors.New("invalid data")

// Signator is used to sign a transaction
/*type Signator interface {
	Sign([]byte) (*keypairs.Signature, error)
	PublicKey() keypairs.PublicKey
	Verify(pubkey, signature, hash []byte) error
}*/

func DefaultRequestOptions() *RequestOptions {
	return &RequestOptions{PeerSetSize: 3}
}

func (m *LogEntryHeader) MarshalBinary() ([]byte, error) {
	tb := make([]byte, 8)
	binary.BigEndian.PutUint64(tb, m.Timestamp)
	th := make([]byte, 8)
	binary.BigEndian.PutUint64(th, m.Height)

	return utils.ConcatByteSlices(tb, m.PrevHash, th, m.Source), nil
}

func (m *LogEntryHeader) UnmarshalBinary(b []byte) error {
	if len(b) < 48 {
		return errInvalidData
	}

	m.Timestamp = binary.BigEndian.Uint64(b[:8])
	m.PrevHash = b[8:40]
	m.Height = binary.BigEndian.Uint64(b[40:48])
	m.Source = b[48:]

	return nil
}

// NewLogEntryBlock instantiates a new LogEntryBlock given the key, previous hash, data
func NewLogEntryBlock(key, prevHash, data []byte) *LogEntryBlock {
	entry := &LogEntryBlock{
		Key: key,
		Header: &LogEntryHeader{
			PrevHash:  prevHash,
			Timestamp: uint64(time.Now().UnixNano()),
		},
		Data: data,
	}
	if utils.IsZeroHash(prevHash) {
		entry.Header.Height = 1
	}
	return entry
}

// EncodeBlock encodes a LogEntryBlock into a Block
func (entry *LogEntryBlock) EncodeBlock() (*Block, error) {
	b, _ := entry.MarshalBinary()
	return &Block{Type: BlockType_LOGENTRY, Data: b}, nil
}

// DecodeBlock decodes a block to a LogEntryBlock
func (entry *LogEntryBlock) DecodeBlock(block *Block) error {
	if block.Type == BlockType_LOGENTRY {
		return entry.UnmarshalBinary(block.Data)
	}
	return ErrInvalidBlockType
}

// DataHash of the tx data
// func (entry *LogEntryBlock) DataHash() []byte {
// 	s := fastsha256.Sum256(entry.Data)
// 	return s[:]
// }

// bytesToGenHash returns the byte slice that should be used to generate the hash
// func (entry *LogEntryBlock) bytesToGenHash() []byte {
// 	headerBytes, _ := entry.Header.MarshalBinary()
// 	return utils.ConcatByteSlices(entry.Key, headerBytes, entry.DataHash())
// }

// ID returns the entry id by hashing the data.
func (entry *LogEntryBlock) ID() []byte {
	// d := entry.bytesToGenHash()
	// s := fastsha256.Sum256(d)
	// return s[:]
	b, _ := entry.EncodeBlock()
	return b.ID()
}

// Sign transaction
/*func (entry *LogEntryBlock) Sign(signer Signator) error {
	entry.Header.Source = signer.PublicKey().Bytes()

	sig, err := signer.Sign(entry.Hash())
	if err == nil {
		entry.Signature = sig.Bytes()
	}

	return err
}

// VerifySignature of the transaction
func (entry *LogEntryBlock) VerifySignature(verifier Signator) error {
	return verifier.Verify(entry.Header.Source, entry.Signature, entry.Hash())
}*/

func (entry LogEntryBlock) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Id":     hex.EncodeToString(entry.ID()),
		"Key":    string(entry.Key),
		"Header": entry.Header,
		"Data":   entry.Data,
	})
}

func (entry *LogEntryBlock) UnmarshalBinary(b []byte) error {
	if len(b) < 10 {
		return errInvalidData
	}

	keySize := binary.BigEndian.Uint16(b[:2])
	ei := 2 + uint64(keySize)
	entry.Key = b[2:ei]

	headerSize := binary.BigEndian.Uint16(b[ei : ei+2])
	ei += 2
	hdr := &LogEntryHeader{}
	if err := hdr.UnmarshalBinary(b[ei : ei+uint64(headerSize)]); err != nil {
		return err
	}
	entry.Header = hdr
	ei += uint64(headerSize)

	dataSize := binary.BigEndian.Uint32(b[ei : ei+4])
	ei += 4
	entry.Data = b[ei : ei+uint64(dataSize)]
	ei += uint64(dataSize)

	sigSize := binary.BigEndian.Uint16(b[ei : ei+2])
	ei += 2
	entry.Signature = b[ei : ei+uint64(sigSize)]
	//ei += uint64(sigSize)

	return nil
}

func (entry *LogEntryBlock) MarshalBinary() ([]byte, error) {
	ksize := make([]byte, 2)
	binary.BigEndian.PutUint16(ksize, uint16(len(entry.Key)))

	headerBytes, _ := entry.Header.MarshalBinary()
	hsize := make([]byte, 2)
	binary.BigEndian.PutUint16(hsize, uint16(len(headerBytes)))

	dsize := make([]byte, 4)
	binary.BigEndian.PutUint32(dsize, uint32(len(entry.Data)))

	ssize := make([]byte, 2)
	binary.BigEndian.PutUint16(ssize, uint16(len(entry.Signature)))

	return utils.ConcatByteSlices(ksize, entry.Key, hsize, headerBytes, dsize, entry.Data, ssize, entry.Signature), nil
}

// NewLogBlock instantiates a new empty Block.
func NewLogBlock(key []byte) *LogBlock {
	return &LogBlock{
		Key:  key,
		Root: make([]byte, 32),
		TXs:  make([][]byte, 0),
	}
}

func (t *LogBlock) EncodeBlock() (*Block, error) {
	d, _ := t.MarshalBinary()
	return &Block{Type: BlockType_LOG, Data: d}, nil
}

func (t *LogBlock) DecodeBlock(blk *Block) error {
	if blk.Type != BlockType_LOG {
		return ErrInvalidBlockType
	}
	return t.UnmarshalBinary(blk.Data)
}

// MarshalBinary marshals the block into a byte slice.
func (t *LogBlock) MarshalBinary() ([]byte, error) {
	tc := make([]byte, 4)
	binary.BigEndian.PutUint32(tc, uint32(len(t.TXs)))
	txs := utils.ConcatByteSlices(t.TXs...)
	return utils.ConcatByteSlices(t.Root, tc, txs, t.Key), nil
}

// UnmarshalBinary unmarshals bytes into a block that were marshalled by MarshalBinary
func (t *LogBlock) UnmarshalBinary(b []byte) error {
	if len(b) < 36 {
		return errInvalidData
	}
	t.Root = b[:32]

	tc := binary.BigEndian.Uint32(b[32:36])
	t.TXs = make([][]byte, tc)
	s := 36
	for i := 0; i < int(tc); i++ {
		t.TXs[i] = b[s : s+32]
		s += 32
	}

	t.Key = b[s:]
	return nil
}

// LastEntry returns the last transaction id in the block.
func (t *LogBlock) LastEntry() []byte {
	l := len(t.TXs)
	if l > 0 {
		return t.TXs[l-1]
	}
	// There is none
	return nil
}

// ContainsEntry returns true if the block contains the tx hash id.
func (t *LogBlock) ContainsEntry(h []byte) bool {
	for _, th := range t.TXs {
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
	if len(t.TXs) > 0 {
		if !utils.EqualBytes(t.TXs[len(t.TXs)-1], entry.Header.PrevHash) {
			return ErrPrevHash
		}
	}

	// Add to block
	t.TXs = append(t.TXs, id)

	// Generate root hash
	sh := fastsha256.Sum256(utils.ConcatByteSlices(t.TXs...))
	t.Root = sh[:]
	//t.Root = merkle.GenerateTree(t.TXs).Root().Hash()
	return nil
}

// Height returns the number of tx's in the block i.e. the txblock height
func (t *LogBlock) Height() int {
	return len(t.TXs)
}

// MarshalJSON custom marshal LogBlock to be human readable.
func (t *LogBlock) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"Key":  string(t.Key),
		"Root": fmt.Sprintf("%x", t.Root),
	}

	s := make([]string, len(t.TXs))
	for i, v := range t.TXs {
		s[i] = fmt.Sprintf("%x", v)
	}
	m["TXs"] = s

	return json.Marshal(m)
}
