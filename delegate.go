package blockring

import (
	"bytes"
	"log"

	chord "github.com/ipkg/go-chord"

	"github.com/hexablock/blockring/rpc"
	"github.com/hexablock/blockring/store"
	"github.com/hexablock/blockring/structs"
	"github.com/hexablock/blockring/utils"
)

type ChordDelegate struct {
	Store *StoreTransport

	ring      *ChordRing
	blockRing *BlockRing

	// Incoming candidate blocks to be locally stored i.e. taken over. These are either stored or
	// forwarded based on location ID
	InBlocks chan *rpc.BlockRPCData

	peerStore store.PeerStore
}

func NewChordDelegate(peerStore store.PeerStore, inBlockBufSize int) *ChordDelegate {
	return &ChordDelegate{
		InBlocks:  make(chan *rpc.BlockRPCData, inBlockBufSize),
		peerStore: peerStore,
	}
}

// Register registers the chord ring to the delegate and starts processing incoming blocks
func (s *ChordDelegate) Register(ring *ChordRing, blockRing *BlockRing) {
	//s.ring = ring
	s.blockRing = blockRing
	s.ring = ring

	go s.startConsuming()
}

func (s *ChordDelegate) acquireBlock(id []byte, loc *structs.Location) {
	// skip if we have the block
	if _, err := s.Store.local.GetBlock(id); err == nil {
		return
	}

	// get the block from the ring.
	_, blk, err := s.blockRing.GetBlock(id)
	if err != nil {
		log.Printf("ERR action=takeover phase=failed block/%x  msg='Failed to get block: %v'", id, err)
	}

	log.Printf("DBG action=takeover phase=begin block/%x dst=%s", id, utils.ShortVnodeID(loc.Vnode))

	// try to set block
	if err = s.Store.SetBlock(loc, blk); err != nil {
		log.Printf("ERR action=takeover phase=failed block/%x dst=%s msg='%v'",
			id, utils.ShortVnodeID(loc.Vnode), err)
	} else {

		log.Printf("DBG action=takeover phase=complete block/%x dst=%s", id, utils.ShortVnodeID(loc.Vnode))
		// TODO:
		// - ReleaseBlock(id)
	}
}

// StartConsuming takes incoming blocks and adds them to the local store if they fall within the perview
// of the host or are transferred to the predecessor.
func (s *ChordDelegate) startConsuming() {
	for b := range s.InBlocks {
		id := b.ID

		_, vn, err := s.ring.LookupHash(id, 1)
		if err != nil {
			log.Println("ERR", err)
			continue
		}

		// takeover block since we own it.
		if vn[0].Host == s.ring.Hostname() {
			s.acquireBlock(id, b.Location)
		} else {
			// re-route
			log.Printf("DBG action=route phase=begin block/%x dst=%s", id, utils.ShortVnodeID(vn[0]))

			loc := &structs.Location{Id: id, Vnode: vn[0]}
			if err = s.Store.remote.TransferBlock(loc, id); err != nil {
				log.Printf("ERR action=route phase=failed block/%x dst=%s msg='%v'", id, utils.ShortVnodeID(vn[0]), err)
			} else {
				log.Printf("DBG action=route phase=complete block/%x dst=%s", id, utils.ShortVnodeID(vn[0]))
			}
		}

	}
}

func (s *ChordDelegate) transferBlocks(local, remote *chord.Vnode) error {
	return s.Store.local.IterBlocks(func(block *structs.Block) error {
		id := block.ID()

		//
		// Handle transferring natural keys.
		//

		// skip blocks that do not belong to the new remote
		if bytes.Compare(id, remote.Id) >= 0 {
			//log.Printf("DBG action=transfer phase=skipping block/%x dst=%s", id, utils.ShortVnodeID(remote))
			return nil
		}

		loc := &structs.Location{Id: id, Vnode: remote}
		log.Printf("DBG action=transfer phase=begin block/%x dst=%s", id, utils.ShortVnodeID(remote))
		if err := s.Store.remote.TransferBlock(loc, id); err != nil {
			log.Printf("ERR action=transfer phase=failed block/%x dst=%s msg='%v'", id, utils.ShortVnodeID(remote), err)
		} else {
			log.Printf("DBG action=transfer phase=complete block/%x dst=%s", id, utils.ShortVnodeID(remote))
		}

		//
		// TODO: handle replicas
		//

		return nil
	})
}

// NewPredecessor is called when a new predecessor is found
func (s *ChordDelegate) NewPredecessor(local, remoteNew, remotePrev *chord.Vnode) {
	log.Printf("INF event=predecessor pred=%s local=%s", utils.ShortVnodeID(remoteNew), utils.ShortVnodeID(local))
	// nothing to do if new predecessor is local.
	if local.Host == remoteNew.Host {
		return
	}
	// add the new peer to our list of known peers
	s.peerStore.AddPeer(remoteNew.Host)

	if err := s.transferBlocks(local, remoteNew); err != nil {
		log.Printf("ERR action=transfer-blocks local=%s remote=%s", utils.ShortVnodeID(local), utils.ShortVnodeID(remoteNew))
	}

}

// Leaving is called when local node is leaving the ring
func (s *ChordDelegate) Leaving(local, pred, succ *chord.Vnode) {
	//log.Printf("DBG [chord] Leaving local=%s succ=%s", shortID(local), shortID(succ))
}

// PredecessorLeaving is called when a predecessor leaves
func (s *ChordDelegate) PredecessorLeaving(local, remote *chord.Vnode) {
	//log.Printf("DBG [chord] PredecessorLeaving local=%s remote=%s", shortID(local), shortID(remote))
}

// SuccessorLeaving is called when a successor leaves
func (s *ChordDelegate) SuccessorLeaving(local, remote *chord.Vnode) {
	//log.Printf("DBG [chord] SuccessorLeaving local=%s remote=%s", shortID(local), shortID(remote))
}

// Shutdown is called when the node is shutting down
func (s *ChordDelegate) Shutdown() {
	log.Println("INF event=shutdown")
}
