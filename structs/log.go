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
	tx := &LogEntryBlock{
		Key: key,
		Header: &LogEntryHeader{
			PrevHash:  prevHash,
			Timestamp: uint64(time.Now().UnixNano()),
		},
		Data: data,
	}
	if utils.IsZeroHash(prevHash) {
		tx.Header.Height = 1
	}
	return tx
}

// EncodeBlock encodes a LogEntryBlock into a Block
func (tx *LogEntryBlock) EncodeBlock() (*Block, error) {
	b, _ := tx.MarshalBinary()
	return &Block{Type: BlockType_LOGENTRYBLOCK, Data: b}, nil
}

// DecodeBlock decodes a block to a LogEntryBlock
func (tx *LogEntryBlock) DecodeBlock(block *Block) error {
	if block.Type == BlockType_LOGENTRYBLOCK {
		return tx.UnmarshalBinary(block.Data)
	}
	return ErrInvalidBlockType
}

// DataHash of the tx data
// func (tx *LogEntryBlock) DataHash() []byte {
// 	s := fastsha256.Sum256(tx.Data)
// 	return s[:]
// }

// bytesToGenHash returns the byte slice that should be used to generate the hash
// func (tx *LogEntryBlock) bytesToGenHash() []byte {
// 	headerBytes, _ := tx.Header.MarshalBinary()
// 	return utils.ConcatByteSlices(tx.Key, headerBytes, tx.DataHash())
// }

// ID returns the entry id by hashing the data.
func (tx *LogEntryBlock) ID() []byte {
	// d := tx.bytesToGenHash()
	// s := fastsha256.Sum256(d)
	// return s[:]
	b, _ := tx.EncodeBlock()
	return b.ID()
}

// Sign transaction
/*func (tx *LogEntryBlock) Sign(signer Signator) error {
	tx.Header.Source = signer.PublicKey().Bytes()

	sig, err := signer.Sign(tx.Hash())
	if err == nil {
		tx.Signature = sig.Bytes()
	}

	return err
}

// VerifySignature of the transaction
func (tx *LogEntryBlock) VerifySignature(verifier Signator) error {
	return verifier.Verify(tx.Header.Source, tx.Signature, tx.Hash())
}*/

// MarshalJSON is a custom JSON marshaller.  It properly formats the hashes
// and includes everything except the transaction data.
func (tx *LogEntryBlock) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Prev":      hex.EncodeToString(tx.Header.PrevHash),
		"Id":        hex.EncodeToString(tx.ID()),
		"Key":       string(tx.Key),
		"Timestamp": tx.Header.Timestamp,
		"Height":    tx.Header.Height,
	})
}

func (tx *LogEntryBlock) UnmarshalBinary(b []byte) error {
	if len(b) < 10 {
		return errInvalidData
	}

	keySize := binary.BigEndian.Uint16(b[:2])
	ei := 2 + uint64(keySize)
	tx.Key = b[2:ei]

	headerSize := binary.BigEndian.Uint16(b[ei : ei+2])
	ei += 2
	hdr := &LogEntryHeader{}
	if err := hdr.UnmarshalBinary(b[ei : ei+uint64(headerSize)]); err != nil {
		return err
	}
	tx.Header = hdr
	ei += uint64(headerSize)

	dataSize := binary.BigEndian.Uint32(b[ei : ei+4])
	ei += 4
	tx.Data = b[ei : ei+uint64(dataSize)]
	ei += uint64(dataSize)

	sigSize := binary.BigEndian.Uint16(b[ei : ei+2])
	ei += 2
	tx.Signature = b[ei : ei+uint64(sigSize)]
	//ei += uint64(sigSize)

	return nil
}

func (tx *LogEntryBlock) MarshalBinary() ([]byte, error) {
	ksize := make([]byte, 2)
	binary.BigEndian.PutUint16(ksize, uint16(len(tx.Key)))

	headerBytes, _ := tx.Header.MarshalBinary()
	hsize := make([]byte, 2)
	binary.BigEndian.PutUint16(hsize, uint16(len(headerBytes)))

	dsize := make([]byte, 4)
	binary.BigEndian.PutUint32(dsize, uint32(len(tx.Data)))

	ssize := make([]byte, 2)
	binary.BigEndian.PutUint16(ssize, uint16(len(tx.Signature)))

	return utils.ConcatByteSlices(ksize, tx.Key, hsize, headerBytes, dsize, tx.Data, ssize, tx.Signature), nil
}

// NewLogBlock instantiates a new empty Block.
func NewLogBlock(key []byte) *LogBlock {
	return &LogBlock{
		Key:  key,
		Root: make([]byte, 32),
		TXs:  make([][]byte, 0),
	}
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
func (t *LogBlock) AppendEntry(tx *LogEntryBlock) error {
	id := tx.ID()
	if len(t.TXs) > 0 {
		if !utils.EqualBytes(t.TXs[len(t.TXs)-1], tx.Header.PrevHash) {
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
