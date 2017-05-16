package blockring

import (
	"bytes"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/hexablock/blockring/store"
	chord "github.com/ipkg/go-chord"
)

func initChordRing(addr string, peers ...string) (*ChordRing, error) {
	conf, err := initConfig(addr)
	if err != nil {
		return nil, err
	}
	conf.Peers = peers
	conf.Chord.StabilizeMin = 10 * time.Millisecond
	conf.Chord.StabilizeMax = 30 * time.Millisecond

	ln, svr, err := newGRPCServer(conf.BindAddr)
	if err != nil {
		return nil, err
	}

	go svr.Serve(ln)

	trans := chord.NewGRPCTransport(svr, conf.Timeouts.RPC, conf.Timeouts.Idle)

	ps := store.NewPeerInMemStore()

	return joinOrBootstrap(conf, ps, trans)
}

func TestChordRingBootstrap(t *testing.T) {
	ring, err := initChordRing("127.0.0.1:43434")
	if err != nil {
		t.Fatal(err)
	}

	<-time.After(100 * time.Millisecond)

	locs, err := ring.LocateReplicatedKey([]byte("key"), 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(locs) != 3 {
		t.Fatal("should have 3")
	}
}

func TestChordRingJoin(t *testing.T) {
	r1, err := initChordRing("127.0.0.1:43435")
	if err != nil {
		t.Fatal(err)
	}
	<-time.After(100 * time.Millisecond)

	r2, err := initChordRing("127.0.0.1:43436", "127.0.0.1:43435")
	if err != nil {
		t.Fatal(err)
	}
	<-time.After(100 * time.Millisecond)

	testkey := []byte("key")

	rk1, _ := r1.LocateReplicatedKey(testkey, 3)
	rk2, _ := r2.LocateReplicatedKey(testkey, 3)

	f1 := false
	for i, rk := range rk1 {
		if !bytes.Equal(rk.Id, rk2[i].Id) {
			t.Fatal("mismatch")
		}

		if rk.Vnode.Host != "127.0.0.1:43436" {
			f1 = true
		}
	}
	if !f1 {
		t.Fatal("all local vnodes")
	}
}

func initConfig(addr string) (*Config, error) {
	c1 := DefaultConfig()
	c1.Chord.StabilizeMin = 3 * time.Second
	c1.Chord.StabilizeMax = 10 * time.Second
	c1.BindAddr = addr
	err := c1.Validate()
	return c1, err
}

func newGRPCServer(addr string) (net.Listener, *grpc.Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, err
	}

	gserver := grpc.NewServer()
	//go gserver.Serve(ln)
	//ctrans := chord.NewGRPCTransport(ln, gserver, 3*time.Second, 300*time.Second)
	return ln, gserver, nil
}
