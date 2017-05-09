package blockring

import "github.com/hexablock/blockring/structs"

type Store interface {
	GetBlock(id []byte) (*structs.Block, error)
	SetBlock(block *structs.Block) error
	IterBlocks(f func(block *structs.Block) error) error
}

type Transport interface {
	GetBlock(loc *structs.Location, id []byte) (*structs.Block, error)
	SetBlock(loc *structs.Location, block *structs.Block) error
	TransferBlock(loc *structs.Location, block *structs.Block) error
	//RegisterStore(Store)
}

// StoreTransport allows to make requests based on Location around the ring.
type StoreTransport struct {
	host   string
	local  Store
	remote Transport
}

func NewStoreTransport(hostname string, local Store, remote Transport) *StoreTransport {
	//remote.RegisterStore(local)
	return &StoreTransport{host: hostname, local: local, remote: remote}
}

func (t *StoreTransport) GetBlock(loc *structs.Location, id []byte) (*structs.Block, error) {
	if loc.Vnode.Host == t.host {
		return t.local.GetBlock(id)
	}

	return t.remote.GetBlock(loc, id)
}

func (t *StoreTransport) SetBlock(loc *structs.Location, block *structs.Block) error {
	if loc.Vnode.Host == t.host {
		return t.local.SetBlock(block)
	}

	return t.remote.SetBlock(loc, block)
}