package store

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/hexablock/blockring/structs"
)

type PeerStore interface {
	Peers() []string
	AddPeer(string) bool
	SetPeers([]*structs.Peer)
}

type PeerInMemStore struct {
	mu    sync.RWMutex
	peers []*structs.Peer
}

func NewPeerInMemStore() *PeerInMemStore {
	return &PeerInMemStore{peers: make([]*structs.Peer, 0)}
}

// Peers returns a slice of all known peers
func (ps *PeerInMemStore) Peers() []string {
	ps.mu.RLock()
	out := make([]string, len(ps.peers))
	for i, p := range ps.peers {
		out[i] = p.Address
	}
	ps.mu.RUnlock()

	return out
}

// AddPeer adds the given peer to the store.  If it exists then the last seen time is updated and false
// is returned
func (ps *PeerInMemStore) AddPeer(peer string) bool {
	ps.mu.RLock()
	for i, p := range ps.peers {
		if peer == p.Address {
			ps.mu.RUnlock()
			ps.mu.Lock()
			ps.peers[i].LastSeen = uint64(time.Now().UnixNano())
			ps.mu.Unlock()
			return false
		}
	}
	ps.mu.RUnlock()

	p := &structs.Peer{Address: peer, LastSeen: uint64(time.Now().UnixNano())}

	ps.mu.Lock()
	ps.peers = append(ps.peers, p)
	ps.mu.Unlock()

	return true
}

// SetPeers updates the peer list with the given peers.
func (ps *PeerInMemStore) SetPeers(peers []*structs.Peer) {
	ps.mu.RLock()
	if peers != nil && len(peers) > 0 {
		ps.mu.RUnlock()

		// TODO: set uniques
		ps.mu.Lock()
		ps.peers = peers
		ps.mu.Unlock()
	}
	ps.mu.RUnlock()
}

type PeerJSONStore struct {
	filename string
	perms    os.FileMode
	*PeerInMemStore
}

func NewPeerJSONStore(filename string) (*PeerJSONStore, error) {
	pj := PeerJSONStore{filename: filename, PeerInMemStore: NewPeerInMemStore(), perms: 0644}

	if _, err := os.Stat(filename); err != nil {
		return &pj, nil
	}

	data, err := ioutil.ReadFile(filename)
	if err == nil {

		if err = json.Unmarshal(data, &pj.peers); err == nil {
			return &pj, nil
		}
	}
	return nil, err
}

func (ps *PeerJSONStore) AddPeer(peer string) bool {
	if ps.PeerInMemStore.AddPeer(peer) {
		ps.Commit()
		return true
	}
	return false
}

// Commit writes the in-memory peer list to the stable store.
func (ps *PeerJSONStore) Commit() error {
	ps.mu.RLock()
	b, err := json.Marshal(ps.peers)
	ps.mu.RUnlock()
	if err == nil {
		err = ioutil.WriteFile(ps.filename, b, ps.perms)
	}
	return err
}
