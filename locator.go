package blockring

import (
	"github.com/hexablock/blockring/structs"
	"github.com/ipkg/difuse/utils"
	chord "github.com/ipkg/go-chord"
)

// Locator implements a location and lookup service
type Locator interface {
	LookupHash(hash []byte, n int) (*chord.Vnode, []*chord.Vnode, error)
	LookupKey(key []byte, n int) ([]byte, *chord.Vnode, []*chord.Vnode, error)
	LocateReplicatedHash(hash []byte, n int) ([]*structs.Location, error)
	LocateReplicatedKey(key []byte, n int) ([]*structs.Location, error)
}

type locatorRouter struct {
	Locator
}

// RouteHash routes a hash around the ring visiting each vnode. It starts by looking up first n vnodes for the id.  If
// the first batch of n vnodes does not contain the block, then the next n vnodes from the last vnode of the
// previous batch is used until a full circle has been made around the ring.
func (rl *locatorRouter) RouteHash(hash []byte, n int, f func(*structs.Location) bool) error {
	lid := hash

	_, vns, err := rl.LookupHash(lid, n)
	if err != nil {
		return err
	}

	// Try the primary vnode
	svn := vns[0]
	pstart := int32(0)

	loc := &structs.Location{Id: hash, Vnode: svn, Priority: pstart}
	if !f(loc) {
		return nil
	}

	// Try successors in n batches
	m := map[string]bool{}
	wset := vns[1:]
	done := false
	pstart++

	for {
		// exclude vnodes we have visited.
		out := make([]*chord.Vnode, 0, n)
		for _, vn := range wset {
			// If we are back at the starting vnode we've completed a full round.
			// Set to exit after this iteration
			if utils.EqualBytes(svn.Id, vn.Id) {
				done = true
				break
			}

			// if we have not visited - track and add to list of vnodes
			if _, ok := m[vn.StringID()]; !ok {
				m[vn.StringID()] = true
				out = append(out, vn)
			}

		}

		for _, vn := range out {
			loc := &structs.Location{Id: hash, Vnode: vn, Priority: pstart}
			if !f(loc) {
				return nil
			}
			pstart++
		}

		if done {
			return nil
		}

		lid = out[len(out)-1].Id

		// update the next set of vnodes to query
		_, vn, err := rl.LookupHash(lid, n)
		if err != nil {
			return err
		}
		wset = vn
	}
}

func (rl *locatorRouter) RouteKey(key []byte, n int, f func(*structs.Location) bool) error {

	keyhash, _, vns, err := rl.LookupKey(key, n)
	if err != nil {
		return err
	}

	lid := keyhash

	// Try the primary vnode
	svn := vns[0]
	pstart := int32(0)

	loc := &structs.Location{Id: keyhash, Vnode: svn, Priority: pstart}
	if !f(loc) {
		return nil
	}

	// Try successors in n batches
	m := map[string]bool{}
	wset := vns[1:]
	done := false
	pstart++

	for {
		// exclude vnodes we have visited.
		out := make([]*chord.Vnode, 0, n)
		for _, vn := range wset {
			// If we are back at the starting vnode we've completed a full round.
			// Set to exit after this iteration
			if utils.EqualBytes(svn.Id, vn.Id) {
				done = true
				break
			}

			// if we have not visited - track and add to list of vnodes
			if _, ok := m[vn.StringID()]; !ok {
				m[vn.StringID()] = true
				out = append(out, vn)
			}

		}

		for _, vn := range out {
			loc := &structs.Location{Id: keyhash, Vnode: vn, Priority: pstart}
			if !f(loc) {
				return nil
			}
			pstart++
		}

		if done {
			return nil
		}

		lid = out[len(out)-1].Id

		// update the next set of vnodes to query
		_, vn, err := rl.LookupHash(lid, n)
		if err != nil {
			return err
		}
		wset = vn
	}
}
