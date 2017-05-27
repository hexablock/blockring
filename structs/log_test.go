package structs

import (
	"bytes"
	"testing"
	"time"
)

var testEntry = &LogEntryBlock{
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

func TestLogBlock(t *testing.T) {
	lb := NewLogBlock([]byte("key"))
	lb.Signature = []byte("signature")
	lb.Header.Source = []byte("source")
	if err := lb.AppendEntry(testEntry); err != nil {
		t.Fatal(err)
	}
	lbb, _ := lb.MarshalBinary()

	var lb2 LogBlock
	if err := lb2.UnmarshalBinary(lbb); err != nil {
		t.Fatal(err)
	}

	if lb2.Header.Timestamp != lb.Header.Timestamp {
		t.Fatal("mismatch")
	}

	if string(lb.Signature) != "signature" {
		t.Error("signature mismatch")
	}

	if string(lb.Header.Source) != "source" {
		t.Error("source mismatch")
	}

	if lb.Height() != lb2.Height() {
		t.Error("hieght mismatch")
	}

	for i := range lb.Entries {
		if bytes.Compare(lb.Entries[i], lb2.Entries[i]) != 0 {
			t.Fatalf("mismatch %d. %x!=%x", i, lb.Entries[i], lb2.Entries[i])
		}
	}
}

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

	b, err := testEntry.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	ent1 := &LogEntryBlock{}
	if err = ent1.UnmarshalBinary(b); err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(ent1.Key, testEntry.Key) != 0 {
		t.Fatal("key mismatch")
	}
	if bytes.Compare(testEntry.Signature, ent1.Signature) != 0 {
		t.Fatal("sig mismatch")
	}
	if bytes.Compare(testEntry.Data, ent1.Data) != 0 {
		t.Fatal("data mismatch")
	}

}
