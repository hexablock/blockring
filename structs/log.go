package structs

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/hexablock/blockring/utils"
)

// ErrPrevHash is an error when the previous hash of an entry does not match
var ErrPrevHash = errors.New("prev hash mismatch")

// DefaultRequestOptions returns a default set of request Options.
func DefaultRequestOptions() *RequestOptions {
	return &RequestOptions{PeerSetSize: 3}
}

// LogEntryHeader contains header information for a LogEntryBlock
type LogEntryHeader struct {
	PrevHash  []byte // hash of previous LogEntryBlock
	Source    []byte // public key of signer
	Timestamp uint64
	Height    uint64 // height in the LogBlock
}

// LogEntryBlock is a single entry on a LogBlock.  It represents a single operation i.e. single unit
type LogEntryBlock struct {
	Key       []byte
	Header    *LogEntryHeader
	Data      []byte
	Signature []byte
}

// Signator is used to sign a transaction
/*type Signator interface {
	Sign([]byte) (*keypairs.Signature, error)
	PublicKey() keypairs.PublicKey
	Verify(pubkey, signature, hash []byte) error
}*/

// MarshalJSON is a custom marshal to handle hashes
func (h LogEntryHeader) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"PrevHash":  hex.EncodeToString(h.PrevHash),
		"Source":    hex.EncodeToString(h.Source),
		"Height":    h.Height,
		"Timestamp": h.Timestamp,
	})
}

// MarshalBinary marshals the LogEntryHeader into a byte slice.
func (h *LogEntryHeader) MarshalBinary() ([]byte, error) {
	tb := make([]byte, 8)
	binary.BigEndian.PutUint64(tb, h.Timestamp)
	th := make([]byte, 8)
	binary.BigEndian.PutUint64(th, h.Height)

	return utils.ConcatByteSlices(tb, h.PrevHash, th, h.Source), nil
}

// UnmarshalBinary unmarshals a the byte slice into a LogEntryHeader
func (h *LogEntryHeader) UnmarshalBinary(b []byte) error {
	if len(b) < 48 {
		return errInvalidBlockData
	}

	h.Timestamp = binary.BigEndian.Uint64(b[:8])
	h.PrevHash = b[8:40]
	h.Height = binary.BigEndian.Uint64(b[40:48])
	h.Source = b[48:]

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

// ID returns the entry id by hashing the data.
func (entry *LogEntryBlock) ID() []byte {
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
		"Id":        hex.EncodeToString(entry.ID()),
		"Key":       string(entry.Key),
		"Header":    entry.Header,
		"Data":      entry.Data,
		"Signature": hex.EncodeToString(entry.Signature),
	})
}

func (entry *LogEntryBlock) UnmarshalBinary(b []byte) error {
	if len(b) < 10 {
		return errInvalidBlockData
	}
	// key
	keySize := binary.BigEndian.Uint16(b[:2])
	ei := 2 + uint64(keySize)
	entry.Key = b[2:ei]
	// header
	headerSize := binary.BigEndian.Uint16(b[ei : ei+2])
	ei += 2
	hdr := &LogEntryHeader{}
	if err := hdr.UnmarshalBinary(b[ei : ei+uint64(headerSize)]); err != nil {
		return err
	}
	entry.Header = hdr
	ei += uint64(headerSize)
	// data
	dataSize := binary.BigEndian.Uint32(b[ei : ei+4])
	ei += 4
	entry.Data = b[ei : ei+uint64(dataSize)]
	ei += uint64(dataSize)
	// signature
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
