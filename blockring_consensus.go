package blockring

import (
	"github.com/hexablock/blockring/structs"
	"github.com/hexablock/log"
)

// ProposeEntry locates participating voters for the LogEntryBlock and submits a propose request to
// each participant, 1 per go-routine
func (br *BlockRing) ProposeEntry(tx *structs.LogEntryBlock, opts structs.RequestOptions) error {

	locs, err := br.locator.LocateReplicatedKey(tx.Key, int(opts.PeerSetSize))
	if err != nil {
		return err
	}

	if opts.Source != nil && len(opts.Source) > 0 {
		// Broadcast to all vnodes skipping the source.
		for _, l := range locs {
			// 1 go-routine per location
			go func(loc *structs.Location) {

				o := structs.RequestOptions{
					Destination: loc.Vnode.Id,
					Source:      opts.Source,
					PeerSetSize: opts.PeerSetSize,
					PeerSetKey:  loc.Id,
				}
				if er := br.logTrans.ProposeEntry(loc, tx, o); er != nil {
					log.Println("[ERROR] Propose", er)
				}

			}(l)

		}

	} else {
		// Broadcast to all vnodes
		for _, l := range locs {

			go func(loc *structs.Location) {

				o := structs.RequestOptions{
					Destination: loc.Vnode.Id,
					Source:      loc.Vnode.Id,
					PeerSetSize: opts.PeerSetSize,
					PeerSetKey:  loc.Id,
				}
				if er := br.logTrans.ProposeEntry(loc, tx, o); er != nil {
					log.Println("[ERROR] Propose", er)
				}

			}(l)

		}

	}

	return nil
}

// CommitEntry locates partipicating voters for the LogEntryBlock and asynchronously submits a commit
// request to each one participant, 1 per go-routine.
func (br *BlockRing) CommitEntry(tx *structs.LogEntryBlock, opts structs.RequestOptions) error {

	locs, err := br.locator.LocateReplicatedKey(tx.Key, int(opts.PeerSetSize))
	if err != nil {
		return err
	}

	if opts.Source != nil && len(opts.Source) > 0 {
		// Broadcast to all vnodes skipping the source.
		for _, l := range locs {
			// 1 go-routine per location
			go func(loc *structs.Location) {

				o := structs.RequestOptions{
					Destination: loc.Vnode.Id,
					Source:      opts.Source,
					PeerSetSize: opts.PeerSetSize,
					PeerSetKey:  loc.Id,
				}
				if er := br.logTrans.CommitEntry(loc, tx, o); er != nil {
					log.Println("[ERROR] Commit", er)
				}

			}(l)

		}

	} else {
		// Broadcast to all vnodes
		for _, l := range locs {

			go func(loc *structs.Location) {

				o := structs.RequestOptions{
					Destination: loc.Vnode.Id,
					Source:      loc.Vnode.Id,
					PeerSetSize: opts.PeerSetSize,
					PeerSetKey:  loc.Id,
				}
				if er := br.logTrans.CommitEntry(loc, tx, o); er != nil {
					log.Println("[ERROR] Commit", er)
				}

			}(l)

		}

	}

	return nil
}
