package netx

import (
	"context"
	"net"
)

type ConnPool struct {
	connCh chan net.Conn
}

func NewConnPool(addr string, size int) (*ConnPool, error) {
	p := &ConnPool{
		connCh: make(chan net.Conn, size),
	}

	for i := 0; i < size; i++ {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		p.connCh <- conn
	}
	return p, nil
}

func (p *ConnPool) Acquire(ctx context.Context) (net.Conn, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case conn := <-p.connCh:
		return conn, nil
	}
}

func (p *ConnPool) Release(conn net.Conn) {
	p.connCh <- conn
}
