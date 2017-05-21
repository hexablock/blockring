package pool

import (
	"sync"
	"time"

	"github.com/hexablock/blockring/rpc"
	"google.golang.org/grpc"
)

// OutConn is an outbound connection
type OutConn struct {
	host      string
	conn      *grpc.ClientConn
	BlockRPC  rpc.BlockRPCClient
	LocateRPC rpc.LocateRPCClient
	LogRPC    rpc.LogRPCClient
	lastUsed  time.Time
}

// OutConnPool is an outbound connection pool
type OutConnPool struct {
	mu   sync.RWMutex
	pool map[string][]*OutConn

	connReapInterval time.Duration // in seconds
	maxIdle          time.Duration // in seconds
}

func NewOutConnPool(reapInterval, maxIdle int) *OutConnPool {
	pool := &OutConnPool{
		pool:             make(map[string][]*OutConn),
		connReapInterval: time.Second * time.Duration(reapInterval),
		maxIdle:          time.Duration(maxIdle) * time.Second,
	}
	go pool.reapOld()
	return pool
}

// Get returns an existing or new connection
func (pool *OutConnPool) Get(host string) (*OutConn, error) {
	// Check if we have a conn cached
	var out *OutConn

	pool.mu.Lock()
	list, ok := pool.pool[host]
	if ok && len(list) > 0 {
		out = list[len(list)-1]
		list = list[:len(list)-1]
		pool.pool[host] = list
	}
	pool.mu.Unlock()

	// Make a new connection
	if out == nil {
		conn, err := grpc.Dial(host, grpc.WithInsecure())
		if err == nil {
			return &OutConn{
				host:      host,
				BlockRPC:  rpc.NewBlockRPCClient(conn),
				LocateRPC: rpc.NewLocateRPCClient(conn),
				LogRPC:    rpc.NewLogRPCClient(conn),
				conn:      conn,
				lastUsed:  time.Now(),
			}, nil
		}
		return nil, err
	}
	// return an existing connection
	return out, nil
}

// Return returns a connection back to the pool
func (pool *OutConnPool) Return(o *OutConn) {
	o.lastUsed = time.Now()
	pool.mu.Lock()

	list, _ := pool.pool[o.host]
	pool.pool[o.host] = append(list, o)
	pool.mu.Unlock()
}

func (pool *OutConnPool) reapOld() {
	for {
		time.Sleep(pool.connReapInterval)
		pool.reapOnce()
	}
}

func (pool *OutConnPool) reapOnce() {
	pool.mu.Lock()

	for host, conns := range pool.pool {
		max := len(conns)
		for i := 0; i < max; i++ {
			if time.Since(conns[i].lastUsed) > pool.maxIdle {
				conns[i].conn.Close()
				conns[i], conns[max-1] = conns[max-1], nil
				max--
				i--
			}
		}

		pool.pool[host] = conns[:max]
	}

	pool.mu.Unlock()
}
