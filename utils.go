package blockring

import (
	"fmt"
	"log"
	"time"

	chord "github.com/ipkg/go-chord"
)

/*func initConfig(addr string) (*Config, error) {
	c1 := DefaultConfig()
	c1.Chord.StabilizeMin = 3 * time.Second
	c1.Chord.StabilizeMax = 10 * time.Second
	c1.BindAddr = addr

	err := c1.ValidateAddrs()
	return c1, err
}

func newGRPCServer(addr string) (net.Listener, *grpc.Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, err
	}

	gserver := grpc.NewServer()
	//go gserver.Serve(ln)
	//ctrans := chord.NewGRPCTransport(ln, gserver, 3*time.Second, 300*time.Second)
	return ln, gserver, nil
}*/

// try joining each peer one by one returning the ring on the first successful join
func joinRing(conf *Config, trans *chord.GRPCTransport) (*ChordRing, error) {

	for _, peer := range conf.Peers {
		log.Printf("Trying peer=%s", peer)
		ring, err := chord.Join(conf.Chord, trans, peer)
		if err == nil {
			return &ChordRing{ring: ring, conf: conf.Chord, trans: trans}, nil
		}
		log.Printf("Failed to connect peer=%s msg='%v'", peer, err)
		<-time.After(1250 * time.Millisecond)
	}

	return nil, fmt.Errorf("all peers exhausted")
}

// RetryJoinRing implements exponential backoff rejoin.
func RetryJoinRing(conf *Config, trans *chord.GRPCTransport) (*ChordRing, error) {

	retryInSec := 2
	tries := 0

	for {
		tries++
		if tries == 3 {
			tries = 0
			retryInSec *= retryInSec
		}
		// try each set of peers
		ring, err := joinRing(conf, trans)
		if err == nil {
			return ring, nil
		}
		log.Printf("Failed to connect msg='%v'", err)
		log.Printf("Trying in %d secs ...", retryInSec)
		<-time.After(time.Duration(retryInSec) * time.Second)

	}

}

// JoinRingOrBootstrap joins or bootstraps based on config.
func JoinRingOrBootstrap(conf *Config, trans *chord.GRPCTransport) (*ChordRing, error) {
	if len(conf.Peers) > 0 {
		ring, err := joinRing(conf, trans)
		if err == nil {
			return ring, nil
		}
		log.Printf("Failed to join msg='%v'", err)
	}

	log.Printf("Starting mode=bootstrap hostname=%s", conf.Chord.Hostname)
	ring, err := chord.Create(conf.Chord, trans)
	if err == nil {
		return &ChordRing{ring: ring, conf: conf.Chord, trans: trans}, nil
	}
	return nil, err
}
