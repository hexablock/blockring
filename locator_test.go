package blockring

import (
	"testing"
	"time"

	"github.com/hexablock/blockring/structs"
)

func Test_locatorRouter(t *testing.T) {
	r1, err := initChordRing("127.0.0.1:53435")
	if err != nil {
		t.Fatal(err)
	}
	<-time.After(100 * time.Millisecond)

	r2, err := initChordRing("127.0.0.1:53436", "127.0.0.1:53435")
	if err != nil {
		t.Fatal(err)
	}
	<-time.After(100 * time.Millisecond)

	r3, err := initChordRing("127.0.0.1:53437", "127.0.0.1:53436")
	if err != nil {
		t.Fatal(err)
	}
	<-time.After(500 * time.Millisecond)

	testkey := []byte("key")

	router := &locatorRouter{Locator: r1, maxSuccessors: r1.conf.NumSuccessors}

	var cnt int
	err = router.RouteKey(testkey, -1, func(loc *structs.Location) bool {
		cnt++
		return true
	})

	if err != nil {
		t.Fatal(err)
	}

	total := r1.conf.NumVnodes + r2.conf.NumVnodes + r3.conf.NumVnodes
	if total != cnt {
		t.Fatalf("all vnodes not traversed: want=%d have=%d", total, cnt)
	}

}
