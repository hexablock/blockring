package client

import (
	"testing"

	"github.com/hexablock/blockring/structs"
	"github.com/ipkg/difuse/utils"
)

func TestClient(t *testing.T) {
	conf := DefaultConfig()
	conf.SetPeers("127.0.0.1:10123")

	client, err := NewClient(conf)
	if err != nil {
		t.Fatal(err)
	}

	blk := structs.NewDataBlock([]byte("foo"))
	if _, err = client.SetBlock(blk); err != nil {
		t.Fatal(err)
	}

	id := blk.ID()
	_, rblk, err := client.GetBlock(id)
	if err != nil {
		t.Fatal(err)
	}

	rid := rblk.ID()
	if !utils.EqualBytes(id, rid) {
		t.Fatal("id's should match")
	}

}
