package blockring

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hexablock/blockring/structs"
	"github.com/hexablock/blockring/utils"
	chord "github.com/ipkg/go-chord"
	"google.golang.org/grpc"
)

// ChordRing contains
type ChordRing struct {
	conf  *chord.Config
	trans *chord.GRPCTransport
	ring  *chord.Ring
}

// NewChordRing instantiates new ChordRing struct joining or creating a ring based on the configuration.
func NewChordRing(conf *Config, gserver *grpc.Server) (*ChordRing, error) {

	var (
		trans = chord.NewGRPCTransport(gserver, conf.Timeouts.RPC, conf.Timeouts.Idle)
		ring  *ChordRing
		err   error
	)

	if len(conf.Peers) > 0 {
		if conf.RetryJoin {
			ring, err = RetryJoinRing(conf, trans)
		} else {
			ring, err = joinRing(conf, trans)
		}
	} else {
		ring, err = JoinRingOrBootstrap(conf, trans)
	}

	if err == nil {
		if dlg, ok := conf.Chord.Delegate.(*ChordDelegate); ok {
			dlg.RegisterRing(ring)
		} else {
			err = errors.New("unsupported chord delegate")
		}
	}

	return ring, err
}

// Vnodes returns all vnodes for a given host.  If host is empty string, local vnodes are returned.
func (cr *ChordRing) Vnodes(host string) ([]*chord.Vnode, error) {
	if host == "" {
		return cr.trans.ListVnodes(cr.conf.Hostname)
	}
	return cr.trans.ListVnodes(host)
}

// LocateKey returns the key hash, pred. vnode, n succesor vnodes
func (cr *ChordRing) LocateKey(key []byte, n int) ([]byte, *chord.Vnode, []*chord.Vnode, error) {
	return cr.ring.Lookup(n, key)
}

// LocateHash returns the pred. vnode, n succesor vnodes
func (cr *ChordRing) LocateHash(hash []byte, n int) (*chord.Vnode, []*chord.Vnode, error) {
	return cr.ring.LookupHash(n, hash)
}

// LocateReplicatedKey returns vnodes where a key and replicas are located.
func (cr *ChordRing) LocateReplicatedKey(key []byte, n int) ([]*structs.Location, error) {
	hashes := utils.ReplicatedKeyHashes(key, n)
	out := make([]*structs.Location, n)

	for i, h := range hashes {
		_, vs, err := cr.ring.LookupHash(1, h)
		if err != nil {
			return nil, err
		}
		out[i] = &structs.Location{Id: h, Vnode: vs[0]}
	}
	return out, nil
}

// LocateReplicatedHash returns vnodes where a key and replicas are located.
func (cr *ChordRing) LocateReplicatedHash(hash []byte, n int) ([]*structs.Location, error) {
	hashes := utils.ReplicaHashes(hash, n)
	out := make([]*structs.Location, n)

	for i, h := range hashes {
		_, vs, err := cr.ring.LookupHash(1, h)
		if err != nil {
			return nil, err
		}
		out[i] = &structs.Location{Id: h, Vnode: vs[0]}
	}
	return out, nil
}

// Hostname returns the hostname of the node per the config.
func (cr *ChordRing) Hostname() string {
	return cr.conf.Hostname
}

// NumSuccessors returns the num of succesors per the config.
func (cr *ChordRing) NumSuccessors() int {
	return cr.conf.NumSuccessors
}

// try joining each peer one by one returning the ring on the first successful join
func joinRing(conf *Config, trans *chord.GRPCTransport) (*ChordRing, error) {

	for _, peer := range conf.Peers {
		log.Printf("Trying peer=%s", peer)
		ring, err := chord.Join(conf.Chord, trans, peer)
		if err == nil {
			return &ChordRing{ring: ring, conf: conf.Chord, trans: trans}, nil
		}
		log.Printf("Failed to connect peer=%s msg='%v'", peer, err)
		<-time.After(1250 * time.Millisecond)
	}

	return nil, fmt.Errorf("all peers exhausted")
}

// RetryJoinRing implements exponential backoff rejoin.
func RetryJoinRing(conf *Config, trans *chord.GRPCTransport) (*ChordRing, error) {

	retryInSec := 2
	tries := 0

	for {
		tries++
		if tries == 3 {
			tries = 0
			retryInSec *= retryInSec
		}
		// try each set of peers
		ring, err := joinRing(conf, trans)
		if err == nil {
			return ring, nil
		}
		log.Printf("Failed to connect msg='%v'", err)
		log.Printf("Trying in %d secs ...", retryInSec)
		<-time.After(time.Duration(retryInSec) * time.Second)

	}

}

// JoinRingOrBootstrap joins or bootstraps based on config.
func JoinRingOrBootstrap(conf *Config, trans *chord.GRPCTransport) (*ChordRing, error) {
	if len(conf.Peers) > 0 {
		ring, err := joinRing(conf, trans)
		if err == nil {
			return ring, nil
		}
		log.Printf("Failed to join msg='%v'", err)
	}

	log.Printf("Starting mode=bootstrap hostname=%s", conf.Chord.Hostname)
	ring, err := chord.Create(conf.Chord, trans)
	if err == nil {
		return &ChordRing{ring: ring, conf: conf.Chord, trans: trans}, nil
	}
	return nil, err
}
