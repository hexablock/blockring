package blockring

import (
	"bytes"
	"log"

	"github.com/hexablock/blockring/rpc"
	"github.com/hexablock/blockring/structs"
	chord "github.com/ipkg/go-chord"

	"github.com/ipkg/difuse/utils"
)

type ChordDelegate struct {
	Store *StoreTransport

	Ring *ChordRing
	// Incoming candidate blocks to be locally stored i.e. taken over. These are either stored or
	// forwarded based on location ID
	InBlocks chan *rpc.BlockRPCData
}

func NewChordDelegate(inBlkBufSize int) *ChordDelegate {
	return &ChordDelegate{InBlocks: make(chan *rpc.BlockRPCData, inBlkBufSize)}
}

// StartConsuming takes incoming blocks and adds them to the local store if they fall within the perview
// of the host or are transferred to the predecessor.
func (s *ChordDelegate) StartConsuming() {
	for b := range s.InBlocks {
		id := b.Block.ID()

		pred, vn, err := s.Ring.LocateHash(id, 1)
		if err != nil {
			log.Println("ERR", err)
			continue
		}
		// re-route to predecessor based on location.Id
		if bytes.Compare(b.Location.Id, pred.Id) < 0 && pred.Host != s.Ring.Hostname() {

			log.Printf("action=route phase=begin block=%x dst=%s", id, utils.ShortVnodeID(pred))
			loc := &structs.Location{Id: id, Vnode: pred}
			if err = s.Store.remote.TransferBlock(loc, b.Block); err != nil {
				log.Printf("ERR action=route phase=failed block=%x dst=%s msg='%v'", id, utils.ShortVnodeID(pred), err)
			} else {
				log.Printf("action=route phase=complete block=%x dst=%s", id, utils.ShortVnodeID(pred))
			}

			continue
		}
		// takeover block since we own it.
		log.Printf("action=takeover phase=begin block=%x dst=%s", id, utils.ShortVnodeID(vn[0]))
		if err = s.Store.SetBlock(b.Location, b.Block); err != nil {
			log.Printf("ERR action=takeover phase=failed block=%x dst=%s msg='%v'",
				id, utils.ShortVnodeID(vn[0]), err)
		} else {
			log.Printf("action=takeover phase=complete block=%x dst=%s", id, utils.ShortVnodeID(vn[0]))
		}

	}
}

func (s *ChordDelegate) transferBlocks(local, remote *chord.Vnode) error {
	return s.Store.local.IterBlocks(func(block *structs.Block) error {
		id := block.ID()

		if bytes.Compare(id, remote.Id) < 0 {
			log.Printf("action=transfer phase=begin block=%x dst=%s", id, utils.ShortVnodeID(remote))

			loc := &structs.Location{Id: id, Vnode: remote}
			if err := s.Store.remote.TransferBlock(loc, block); err != nil {
				log.Printf("action=transfer phase=failed block=%x dst=%s msg='%v'", id, utils.ShortVnodeID(remote), err)
			} else {
				log.Printf("action=transfer phase=complete block=%x dst=%s", id, utils.ShortVnodeID(remote))
			}

		}

		return nil
	})
}

// NewPredecessor is called when a new predecessor is found
func (s *ChordDelegate) NewPredecessor(local, remoteNew, remotePrev *chord.Vnode) {
	log.Printf("INF action=predecessor pred=%s local=%s", utils.ShortVnodeID(remoteNew), utils.ShortVnodeID(local))

	if local.Host == remoteNew.Host {
		return
	}

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
	log.Println("[chord] Shutdown")
}
