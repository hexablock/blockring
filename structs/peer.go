package structs

// RingPeer is a single peer on the ring.
type Peer struct {
	Address  string
	LastSeen uint64
}
