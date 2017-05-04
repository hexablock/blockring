package utils

import (
	"fmt"
	"testing"

	"github.com/btcsuite/fastsha256"
)

func Test_ReplicatedKeyHashes(t *testing.T) {
	hashes := ReplicatedKeyHashes([]byte("key"), 3)
	if len(hashes) != 3 {
		t.Fatal("should have 3 hashes")
	}

	for _, v := range hashes {
		fmt.Printf("%x\n", v)
	}
}

func Test_ReplicaHashes(t *testing.T) {
	sh := fastsha256.Sum256([]byte("key"))
	hashes := ReplicaHashes(sh[:], 3)
	if len(hashes) != 3 {
		t.Fatal("should have 3 hashes")
	}

	for _, v := range hashes {
		fmt.Printf("%x\n", v)
	}
}
