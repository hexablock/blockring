package blockring

import (
	"errors"
	"fmt"
	"time"

	"github.com/btcsuite/fastsha256"
	"github.com/hexablock/log"

	"github.com/hexablock/blockring/store"
	"github.com/hexablock/blockring/structs"
	"github.com/hexablock/blockring/utils"
	chord "github.com/ipkg/go-chord"
	"google.golang.org/grpc"
)

var errNoPeersFound = errors.New("no peers found")

// ChordRing contains
type ChordRing struct {
	conf      *chord.Config
	trans     *chord.GRPCTransport
	ring      *chord.Ring
	peerStore store.PeerStore
}

// NewChordRing instantiates new ChordRing struct joining or creating a ring based on the configuration.
func NewChordRing(conf *Config, peerStore store.PeerStore, gserver *grpc.Server) (*ChordRing, error) {

	var (
		trans = chord.NewGRPCTransport(gserver, conf.Timeouts.RPC, conf.Timeouts.Idle)
		ring  *ChordRing
		err   error
	)

	if conf.RetryJoin {
		ring, err = retryJoinRing(conf, peerStore, trans)
	} else {
		ring, err = joinOrBootstrap(conf, peerStore, trans)
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

// LookupKey returns the key hash, pred. vnode, n succesor vnodes
func (cr *ChordRing) LookupKey(key []byte, n int) ([]byte, *chord.Vnode, []*chord.Vnode, error) {
	return cr.ring.Lookup(n, key)
}

// LookupHash returns the pred. vnode, n succesor vnodes
func (cr *ChordRing) LookupHash(hash []byte, n int) (*chord.Vnode, []*chord.Vnode, error) {
	return cr.ring.LookupHash(n, hash)
}

// LocateReplicatedKey returns vnodes where a key and replicas are located.
func (cr *ChordRing) LocateReplicatedKey(key []byte, n int) ([]*structs.Location, error) {
	// hashes := utils.ReplicatedKeyHashes(key, n)
	// out := make([]*structs.Location, n)
	//
	// for i, h := range hashes {
	// 	_, vs, err := cr.ring.LookupHash(cr.conf.NumSuccessors, h)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	out[i] = &structs.Location{Id: h, Vnode: vs[0]}
	// }
	// return out, nil
	hash := fastsha256.Sum256(key)
	return cr.LocateReplicatedHash(hash[:], n)
}

// LocateReplicatedHash returns vnodes where a key and replicas are located.
func (cr *ChordRing) LocateReplicatedHash(hash []byte, n int) ([]*structs.Location, error) {
	hashes := utils.ReplicaHashes(hash, n)
	//out := make([]*structs.Location, n)
	out := map[string]*structs.Location{}

	for _, h := range hashes {
		// Lookup first replicated hash
		_, vs, err := cr.ring.LookupHash(cr.conf.NumSuccessors, h)
		if err != nil {
			return nil, err
		}
		// If we already have the current host location, go through the successors to find the
		// next best location
		if _, ok := out[vs[0].Host]; !ok {
			out[string(h)] = &structs.Location{Id: h, Vnode: vs[0]}
		} else {
			for _, vn := range vs[1:] {
				if _, ok := out[vn.Host]; !ok {
					out[string(h)] = &structs.Location{Id: h, Vnode: vn}
					break
				}
			}
		}

	}

	if len(out) != n {
		return nil, fmt.Errorf("not enough hosts found")
	}

	o := make([]*structs.Location, n)
	for i, h := range hashes {
		o[i] = out[string(h)]
	}

	return o, nil
}

// Hostname returns the hostname of the node per the config.
func (cr *ChordRing) Hostname() string {
	return cr.conf.Hostname
}

// NumSuccessors returns the num of succesors per the config.
func (cr *ChordRing) NumSuccessors() int {
	return cr.conf.NumSuccessors
}

// join ring if peers provided.
func joinOrBootstrap(conf *Config, peerStore store.PeerStore, trans *chord.GRPCTransport) (*ChordRing, error) {

	if len(conf.Peers) > 0 {

		if conf.RetryJoin {
			return retryJoinRing(conf, peerStore, trans)
		}

		// join
		_, ring, err := joinRing(conf, peerStore, trans)
		if err == nil {
			return ring, nil
		}
	}

	// if we have known peers and we have not left the ring then simply start the ring as other members
	// will assimilate.  not doing so at times causes a panic.
	if len(peerStore.Peers()) > 0 {
		log.Printf("Starting hostname=%s", conf.Chord.Hostname)
	} else {
		log.Printf("Starting mode=bootstrap hostname=%s", conf.Chord.Hostname)
	}

	// create
	cring, err := chord.Create(conf.Chord, trans)
	if err == nil {
		return &ChordRing{ring: cring, conf: conf.Chord, trans: trans, peerStore: peerStore}, nil
	}
	return nil, err
}

// try joining each peer one by one returning the peer and ring on the first successful join
func joinRing(conf *Config, peerStore store.PeerStore, trans *chord.GRPCTransport) (string, *ChordRing, error) {
	knownPeers := peerStore.Peers()
	peers := dedup(append(conf.Peers, knownPeers...))

	for _, peer := range peers {
		log.Printf("Trying peer=%s", peer)
		ring, err := chord.Join(conf.Chord, trans, peer)
		if err == nil {
			cr := &ChordRing{ring: ring, conf: conf.Chord, trans: trans, peerStore: peerStore}
			cr.peerStore.AddPeer(peer)
			return peer, cr, nil
		}

		log.Printf("Failed to connect peer=%s msg='%v'", peer, err)
		peerStore.RemovePeer(peer)

		<-time.After(1250 * time.Millisecond)
	}

	return "", nil, fmt.Errorf("all peers exhausted")
}

// RetryJoinRing implements exponential backoff rejoin.
func retryJoinRing(conf *Config, peerStore store.PeerStore, trans *chord.GRPCTransport) (*ChordRing, error) {

	retryInSec := 2
	tries := 0

	for {
		tries++
		if tries == 3 {
			tries = 0
			retryInSec *= retryInSec
		}
		// try each set of peers
		_, ring, err := joinRing(conf, peerStore, trans)
		if err == nil {
			return ring, nil
		}
		log.Printf("Failed to connect msg='%v'", err)
		log.Printf("Trying in %d secs ...", retryInSec)
		<-time.After(time.Duration(retryInSec) * time.Second)

	}

}

func dedup(list []string) []string {
	m := map[string]int{}
	for i, l := range list {
		m[l] = i
	}

	o := make([]string, len(m))
	i := 0
	for k := range m {
		o[i] = k
		i++
	}
	return o
}
