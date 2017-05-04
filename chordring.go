package blockring

import (
	"github.com/hexablock/blockring/structs"
	"github.com/hexablock/blockring/utils"
	chord "github.com/ipkg/go-chord"
)

// ChordRing contains
type ChordRing struct {
	conf  *chord.Config
	trans *chord.GRPCTransport
	ring  *chord.Ring
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
