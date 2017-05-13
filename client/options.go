package client

const defaultPeerRange = 3

type RequestOptions struct {
	// Number of replicas for data set
	Replicas int
	// Number of successors to ask when key not found
	PeerRange int
}

func DefaultRequestOptions() *RequestOptions {
	return &RequestOptions{Replicas: 1, PeerRange: defaultPeerRange}
}
