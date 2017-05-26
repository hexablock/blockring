package blockring

import (
	"bytes"

	"github.com/btcsuite/fastsha256"
	"github.com/hexablock/log"

	chord "github.com/ipkg/go-chord"

	"github.com/hexablock/blockring/rpc"
	"github.com/hexablock/blockring/store"
	"github.com/hexablock/blockring/structs"
	"github.com/hexablock/blockring/utils"
)

type ChordDelegate struct {
	Store    *BlockRingTransport
	LogTrans *LogRingTransport

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
	s.ring = ring
	s.blockRing = blockRing

	go s.startConsuming()
}

func (s *ChordDelegate) takeoverLogBlock(key []byte, loc *structs.Location) {
	// get block from network
	opts := structs.RequestOptions{PeerSetKey: loc.Id, PeerSetSize: int32(s.ring.NumSuccessors())}
	blk, _, err := s.LogTrans.GetLogBlock(loc, key, opts)
	if err != nil {
		log.Printf("[ERROR] action=takeover phase=failed key/%s  msg='Failed to get log block: %v'", key, err)
		return
	}

	s.LogTrans.txl.Replay(blk)
}

func (s *ChordDelegate) takeoverBlock(id []byte, loc *structs.Location) {
	// skip if we have the block
	if _, err := s.Store.local.GetBlock(id); err == nil {
		return
	}

	// get the block from the ring.
	_, blk, err := s.blockRing.GetBlock(id)
	if err != nil {
		log.Printf("[ERROR] action=takeover phase=failed block/%x  msg='Failed to get block: %v'", id, err)
		return
	}

	log.Printf("[DEBUG] action=takeover phase=begin block/%x dst=%s", id, utils.ShortVnodeID(loc.Vnode))
	if err = s.Store.SetBlock(loc, blk); err != nil {
		log.Printf("[ERROR] action=takeover phase=failed block/%x dst=%s msg='%v'",
			id, utils.ShortVnodeID(loc.Vnode), err)
	} else {
		log.Printf("[DEBUG] action=takeover phase=complete block/%x dst=%s", id, utils.ShortVnodeID(loc.Vnode))
	}
}

func (s *ChordDelegate) takeoverOrRouteLogBlock(brd *rpc.BlockRPCData) error {
	key := brd.ID
	locs, err := s.ring.LocateReplicatedKey(key, 1)

	if err != nil {
		return err
	}

	loc := locs[0]
	if loc.Vnode.Host == s.ring.Hostname() {
		s.takeoverLogBlock(key, loc)
	} else {
		// re-route
		log.Printf("[DEBUG] action=route phase=begin key/%s dst=%s", key, utils.ShortVnodeID(loc.Vnode))
		if _, err = s.LogTrans.TransferLogBlock(loc, key, structs.RequestOptions{}); err != nil {
			log.Printf("[ERROR] action=route phase=failed key/%s dst=%s msg='%v'", key, utils.ShortVnodeID(loc.Vnode), err)
		} else {
			log.Printf("[DEBUG] action=route phase=complete key/%s dst=%s", key, utils.ShortVnodeID(loc.Vnode))
		}
	}
	return nil
}

func (s *ChordDelegate) takeoverOrRouteBlock(b *rpc.BlockRPCData) error {
	id := b.ID
	locs, err := s.ring.LocateReplicatedHash(id, 1)
	if err != nil {
		return err
	}

	loc := locs[0]
	if loc.Vnode.Host == s.ring.Hostname() {
		s.takeoverBlock(id, b.Location)
	} else {
		// re-route
		log.Printf("[DEBUG] action=route phase=begin block/%x dst=%s", id, utils.ShortVnodeID(loc.Vnode))
		if err = s.Store.remote.TransferBlock(loc, id); err != nil {
			log.Printf("[ERROR] action=route phase=failed block/%x dst=%s msg='%v'", id, utils.ShortVnodeID(loc.Vnode), err)
		} else {
			log.Printf("[DEBUG] action=route phase=complete block/%x dst=%s", id, utils.ShortVnodeID(loc.Vnode))
		}
	}
	return nil
}

// StartConsuming takes incoming blocks and adds them to the local store if they fall within the perview
// of the host or are transferred to the predecessor.
func (s *ChordDelegate) startConsuming() {
	for b := range s.InBlocks {

		var err error
		if b.Block != nil && b.Block.Type == structs.BlockType_LOG {
			err = s.takeoverOrRouteLogBlock(b)
		} else {
			err = s.takeoverOrRouteBlock(b)
		}

		if err != nil {
			log.Println("[ERROR]", err)
		}

	}
}

func (s *ChordDelegate) transferLogBlocks(local, remote *chord.Vnode) error {
	return s.LogTrans.bs.IterKeys(func(key []byte) error {
		//
		// Handle transferring natural keys.
		//

		sh := fastsha256.Sum256(key)
		// skip blocks that do not belong to the new remote
		if bytes.Compare(sh[:], remote.Id) >= 0 {
			return nil
		}

		loc := &structs.Location{Id: sh[:], Vnode: remote}
		log.Printf("[DEBUG] action=transfer phase=begin key=%s dst=%s", key, utils.ShortVnodeID(remote))
		if _, err := s.LogTrans.TransferLogBlock(loc, key, structs.RequestOptions{}); err != nil {
			log.Printf("[ERROR] action=transfer phase=failed key=%s dst=%s msg='%v'", key, utils.ShortVnodeID(remote), err)
		} else {
			log.Printf("[DEBUG] action=transfer phase=complete key=%s dst=%s", key, utils.ShortVnodeID(remote))
		}

		//
		// TODO: handle replicas
		//

		return nil
	})
}

// transfer all but LogBlocks
func (s *ChordDelegate) transferBlocks(local, remote *chord.Vnode) error {
	return s.Store.local.IterIDs(func(id []byte) error {
		//
		// Handle transferring natural keys.
		//

		// skip blocks that do not belong to the new remote
		if bytes.Compare(id, remote.Id) >= 0 {
			return nil
		}

		loc := &structs.Location{Id: id, Vnode: remote}
		log.Printf("[DEBUG] action=transfer phase=begin block=%x dst=%s", id, utils.ShortVnodeID(remote))
		if err := s.Store.remote.TransferBlock(loc, id); err != nil {
			log.Printf("[ERROR] action=transfer phase=failed block=%x dst=%s msg='%v'", id, utils.ShortVnodeID(remote), err)
		} else {
			log.Printf("[DEBUG] action=transfer phase=complete block=%x dst=%s", id, utils.ShortVnodeID(remote))
		}

		//
		// TODO: handle replicas
		//

		return nil
	})
}

// NewPredecessor is called when a new predecessor is found
func (s *ChordDelegate) NewPredecessor(local, remoteNew, remotePrev *chord.Vnode) {
	log.Printf("[INFO] event=predecessor pred=%s local=%s", utils.ShortVnodeID(remoteNew), utils.ShortVnodeID(local))
	// nothing to do if new predecessor is local.
	if local.Host == remoteNew.Host {
		return
	}
	// add the new peer to our list of known peers
	s.peerStore.AddPeer(remoteNew.Host)

	if err := s.transferLogBlocks(local, remoteNew); err != nil {
		log.Printf("[ERROR] action=transfer-blocks local=%s remote=%s", utils.ShortVnodeID(local), utils.ShortVnodeID(remoteNew))
	}

	if err := s.transferBlocks(local, remoteNew); err != nil {
		log.Printf("[ERROR] action=transfer-blocks local=%s remote=%s", utils.ShortVnodeID(local), utils.ShortVnodeID(remoteNew))
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
	log.Println("[INFO] event=shutdown")
}
