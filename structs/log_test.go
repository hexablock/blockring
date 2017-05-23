package structs

import (
	"bytes"
	"testing"
	"time"
)

func TestHeader(t *testing.T) {
	h1 := &LogEntryHeader{
		PrevHash:  make([]byte, 32),
		Height:    1,
		Timestamp: uint64(time.Now().UnixNano()),
		Source:    make([]byte, 40),
	}

	b, _ := h1.MarshalBinary()

	h2 := &LogEntryHeader{}
	if err := h2.UnmarshalBinary(b); err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(h1.PrevHash, h2.PrevHash) != 0 {
		t.Fatal("prev mismatch")
	}
	if bytes.Compare(h1.Source, h2.Source) != 0 {
		t.Fatal("source mismatch")
	}
	if h1.Timestamp != h2.Timestamp {
		t.Fatal("timestamp mismatch")
	}
	if h1.Height != h2.Height {
		t.Fatal("height mismatch")
	}

}

func TestEntry(t *testing.T) {

	ent := &LogEntryBlock{
		Key: []byte("my-test-key"),
		Header: &LogEntryHeader{
			PrevHash:  make([]byte, 32),
			Height:    1,
			Timestamp: uint64(time.Now().UnixNano()),
			Source:    make([]byte, 40),
		},
		Signature: []byte("test-signature"),
		Data:      []byte("some-test-data"),
	}

	b, err := ent.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	ent1 := &LogEntryBlock{}
	if err = ent1.UnmarshalBinary(b); err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(ent1.Key, ent.Key) != 0 {
		t.Fatal("key mismatch")
	}
	if bytes.Compare(ent.Signature, ent1.Signature) != 0 {
		t.Fatal("sig mismatch")
	}
	if bytes.Compare(ent.Data, ent1.Data) != 0 {
		t.Fatal("data mismatch")
	}

}
