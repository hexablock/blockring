package blockring

import (
	"fmt"
	"hash"
	"net"
	"strings"
	"time"

	"github.com/btcsuite/fastsha256"
	chord "github.com/ipkg/go-chord"

	"github.com/hexablock/blockring/utils"
)

// Config holds the overall config
type Config struct {
	Chord *chord.Config

	BindAddr string // Bind address
	AdvAddr  string // Advertise address used by peers

	Peers     []string // Existing peers to join
	RetryJoin bool     // keep trying peers on failure

	Timeouts       *NetTimeouts
	InBlockBufSize int
}

// DefaultConfig returns a sane config
func DefaultConfig() *Config {
	c := &Config{
		Chord:    chord.DefaultConfig(""),
		Timeouts: DefaultNetTimeouts(),
	}

	c.Chord.NumSuccessors = 7
	c.Chord.NumVnodes = 3
	c.Chord.StabilizeMin = 3 * time.Second
	c.Chord.StabilizeMax = 8 * time.Second
	c.Chord.HashFunc = func() hash.Hash { return fastsha256.New() }

	c.InBlockBufSize = c.Chord.NumSuccessors * c.Chord.NumVnodes

	return c
}

// SetPeers parses a comma separated list of peers into a slice and sets the config.
func (cfg *Config) SetPeers(peers string) {
	cfg.Peers = utils.StringToSlice(peers, ",")
}

// Validate validates the config
func (cfg *Config) Validate() error {
	return cfg.validateAddrs()
}

// ValidateAddrs validates both bind and adv addresses and set adv if possible.
func (cfg *Config) validateAddrs() error {
	if _, err := net.ResolveTCPAddr("tcp4", cfg.BindAddr); err != nil {
		return err
	}

	// Check if adv addr is set and validate
	if len(cfg.AdvAddr) != 0 {
		isAdv, err := utils.IsAdvertisableAddress(cfg.AdvAddr)
		if err != nil {
			return err
		} else if !isAdv {
			return fmt.Errorf("address not advertisable")
		}
	} else {
		// Use bind addr if possible
		if isAdv, _ := utils.IsAdvertisableAddress(cfg.BindAddr); isAdv {
			cfg.AdvAddr = cfg.BindAddr
		} else {
			// Lastly, try to auto-detect the ip
			ips, err := utils.AutoDetectIPAddress()
			if err != nil {
				return err
			} else if len(ips) == 0 {
				return fmt.Errorf("could not get advertise address")
			}
			// Add port to advertise address based the one supplied in the bind address
			pp := strings.Split(cfg.BindAddr, ":")
			cfg.AdvAddr = ips[0] + ":" + pp[len(pp)-1]
		}
	}

	cfg.Chord.Hostname = cfg.AdvAddr

	return nil
}

// NetTimeouts holds timeouts for rpc's
type NetTimeouts struct {
	Dial time.Duration
	RPC  time.Duration
	Idle time.Duration
}

// DefaultNetTimeouts initializes sane timeouts
func DefaultNetTimeouts() *NetTimeouts {
	return &NetTimeouts{
		Dial: 3 * time.Second,
		RPC:  5 * time.Second,
		Idle: 300 * time.Second,
	}
}
