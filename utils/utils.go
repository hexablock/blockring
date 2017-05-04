package utils

import (
	"errors"
	"fmt"
	"math/big"
	"net"
	"strings"

	"github.com/btcsuite/fastsha256"
	chord "github.com/ipkg/go-chord"
)

var keyspaceSize = new(big.Int).SetBytes([]byte{
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
	255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255,
})

// ReplicatedKeyHashes calculates all hashes for a replicated key.  It uses  the method mentioned in
// this paper: https://arxiv.org/pdf/1006.3465.pdf
func ReplicatedKeyHashes(key []byte, r int) [][]byte {

	sh := fastsha256.Sum256(key)

	out := make([][]byte, 0, r)
	out = append(out, sh[:])

	rsp := CalculateReplicaHashes(out[0], r)
	for _, v := range rsp {
		out = append(out, v.Bytes())
	}
	return out
}

// ReplicaHashes returns all replica hashes including the startHash
func ReplicaHashes(startHash []byte, r int) [][]byte {
	out := make([][]byte, r)
	out[0] = startHash

	rsp := CalculateReplicaHashes(startHash, r)
	for i, v := range rsp {
		out[i+1] = v.Bytes()
	}
	return out
}

// CalculateReplicaHashes computes r-1 additional hashes around the ring for the given hash and returns
// the respective int values.
func CalculateReplicaHashes(hash []byte, r int) []*big.Int {
	replicas := big.NewInt(int64(r))
	width := new(big.Int).Div(keyspaceSize, replicas)

	ki := new(big.Int).SetBytes(hash)

	out := make([]*big.Int, r-1)

	for i := 1; i < r; i++ {
		bi := big.NewInt(int64(i))
		ti := new(big.Int).Mul(bi, width)
		ti.Add(ki, ti)

		out[i-1] = new(big.Int).Mod(ti, keyspaceSize)
	}

	return out
}

// ZeroHash returns a 32 byte hash with all zeros
func ZeroHash() []byte {
	return make([]byte, 32)
}

// EqualBytes returns whether 2 byte slices have the same data
func EqualBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// IsZeroHash returns whether the hash is zero
func IsZeroHash(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}

// ShortVnodeID returns the shortened vnode id
func ShortVnodeID(vn *chord.Vnode) string {
	if vn == nil {
		return "<nil>/<nil>"
	}
	return fmt.Sprintf("%s/%x", vn.Host, vn.Id[:8])
}

// LongVnodeID returns the full vnode id
func LongVnodeID(vn *chord.Vnode) string {
	if vn == nil {
		return "<nil>/<nil>"
	}
	return fmt.Sprintf("%s/%x", vn.Host, vn.Id)
}

// AutoDetectIPAddress traverses interfaces eliminating, localhost, ifaces with
// no addresses and ipv6 addresses.  It returns a list by priority
func AutoDetectIPAddress() ([]string, error) {

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	out := []string{}

	for _, ifc := range ifaces {
		if strings.HasPrefix(ifc.Name, "lo") {
			continue
		}
		addrs, err := ifc.Addrs()
		if err != nil || len(addrs) == 0 {
			continue
		}

		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}

			// Add ipv4 addresses to list
			if ip.To4() != nil {
				out = append(out, ip.String())
			}
		}

	}
	if len(out) == 0 {
		return nil, fmt.Errorf("could not detect ip addresses")
	}

	return out, nil
}

// IsAdvertisableAddress checks if an address can be used as the advertise address.
func IsAdvertisableAddress(hp string) (bool, error) {
	if hp == "" {
		return false, fmt.Errorf("invalid address")
	}

	pp := strings.Split(hp, ":")
	if len(pp) < 1 {
		return false, fmt.Errorf("could not parse: %s", hp)
	}

	if pp[0] == "" {
		return false, nil
	} else if pp[0] == "0.0.0.0" {
		return false, nil
	}

	if _, err := net.ResolveIPAddr("ip4", pp[0]); err != nil {
		return false, err
	}

	return true, nil
}

// MergeErrors merges 2 errors into a single error separated by a ;
func MergeErrors(err1, err2 error) error {
	if err1 == nil {
		return err2
	} else if err2 == nil {
		return err1
	}
	return errors.New(err1.Error() + ";" + err2.Error())
}

// ConcatByteSlices concatenates byte slices into a single slices.
func ConcatByteSlices(pieces ...[]byte) []byte {
	sz := 0
	for _, p := range pieces {
		sz += len(p)
	}

	buf := make([]byte, sz)

	i := 0
	for _, p := range pieces {
		copy(buf[i:], p)
		i += len(p)
	}
	return buf
}

// StringToSlice parse a delim delimited string to a slice of strings excluding empty strings.
func StringToSlice(s, delim string) []string {
	out := []string{}
	for _, v := range strings.Split(s, delim) {
		if c := strings.TrimSpace(v); c != "" {
			out = append(out, c)
		}
	}
	return out
}
