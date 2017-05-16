package pool

import (
	"sync"
	"time"

	"github.com/hexablock/blockring/rpc"
	"google.golang.org/grpc"
)

type OutConn struct {
	host      string
	conn      *grpc.ClientConn
	BlockRPC  rpc.BlockRPCClient
	LocateRPC rpc.LocateRPCClient
	lastUsed  int64
}

type OutConnPool struct {
	mu  sync.RWMutex
	out map[string]*OutConn

	connReapInterval time.Duration // in seconds
	maxIdle          int           // in seconds
}

func NewOutConnPool(reapInterval, maxIdle int) *OutConnPool {
	pool := &OutConnPool{
		out:              make(map[string]*OutConn),
		connReapInterval: time.Second * time.Duration(reapInterval),
		maxIdle:          maxIdle,
	}
	go pool.reapConns()
	return pool
}

func (pool *OutConnPool) Return(conn *OutConn) {
	pool.mu.RLock()
	if _, ok := pool.out[conn.host]; ok {
		pool.mu.RUnlock()
		pool.mu.Lock()
		pool.out[conn.host].lastUsed = time.Now().Unix()
		pool.mu.Unlock()
	} else {
		pool.mu.Unlock()
	}
}

func (pool *OutConnPool) Get(host string) (*OutConn, error) {
	pool.mu.RLock()
	if v, ok := pool.out[host]; ok {
		defer pool.mu.RUnlock()
		return v, nil
	}
	pool.mu.RUnlock()

	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	oc := &OutConn{
		host:      host,
		conn:      conn,
		BlockRPC:  rpc.NewBlockRPCClient(conn),
		LocateRPC: rpc.NewLocateRPCClient(conn),
		lastUsed:  time.Now().Unix(),
	}

	pool.mu.Lock()
	pool.out[host] = oc
	pool.mu.Unlock()

	return oc, nil

}

// reapConns reaps idle connections every 30 seconds
func (pool *OutConnPool) reapConns() {
	for {
		//time.Sleep(30 * time.Second)
		time.Sleep(pool.connReapInterval)
		pool.reapOnce()
	}
}

// reapOnce does a one time reap across all connections, closing and removing any that have been idle
// for more than 2 mins.
func (pool *OutConnPool) reapOnce() {

	now := time.Now().Unix()

	pool.mu.Lock()
	for k, v := range pool.out {
		// expire all older than maxIdleInSecs
		if (now - v.lastUsed) > int64(pool.maxIdle) {
			v.conn.Close()
			delete(pool.out, k)
		}
	}
	pool.mu.Unlock()
}
