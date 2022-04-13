package netx

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
)

// Conn is a stream oriented network connection with i/o operations that are controlled by Contexts
type CtxConn interface {
	net.Conn
	ReadContext(ctx context.Context, b []byte) (n int, err error)
	WriteContext(ctx context.Context, b []byte) (n int, err error)
}

// Conn is a generic stream-oriented network connection with stats.
//
// Only Read & Write methods contain additional logic for stats.
//
// See: TCPListener or UDPListener how to create it.
type Conn struct {
	net.Conn
	stats     *Stats
	closeOnce sync.Once
}

// ReadContext does same as Read method but with a context.
// This method requires 1 additional goroutine from a worker pool.
func (c *Conn) ReadContext(ctx context.Context, b []byte) (n int, err error) {
	// TODO: bytes pool
	buf := make([]byte, len(b))
	// TODO: chan pool
	ch := make(chan ioResult, 1)

	// TODO: goroutine pool
	go func() { ch <- newIOResult(c.Read(buf)) }()

	select {
	case res := <-ch:
		copy(b, buf)
		return res.n, res.err
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

// WriteContext does same as Write method but with a context.
// This method requires 1 additional goroutine from a worker pool.
func (c *Conn) WriteContext(ctx context.Context, b []byte) (n int, err error) {
	// TODO: bytes pool
	buf := make([]byte, len(b))
	copy(buf, b)
	// TODO: chan pool
	ch := make(chan ioResult, 1)

	// TODO: goroutine pool
	go func() { ch <- newIOResult(c.Write(buf)) }()

	select {
	case res := <-ch:
		return res.n, res.err
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

// Read reads data from the connection.
// Read can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetReadDeadline.
func (c *Conn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	c.stats.readBytesAdd(n)
	if err != nil && err != io.EOF {
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			c.stats.readTimeoutsInc()
		} else {
			c.stats.readErrorsInc()
		}
	}
	return n, err
}

// Write writes data to the connection.
// Write can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetWriteDeadline.
func (c *Conn) Write(p []byte) (int, error) {
	n, err := c.Conn.Write(p)
	c.stats.writtenBytesAdd(n)
	if err != nil {
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			c.stats.writeTimeoutsInc()
		} else {
			c.stats.writeErrorsInc()
		}
	}
	return n, err
}

func (c *Conn) Close() error {
	var err error
	c.closeOnce.Do(func() {
		err = c.Conn.Close()
		c.stats.connsInc()
		if err != nil {
			c.stats.closeErrorsInc()
		}
	})
	return err
}

type ioResult struct {
	n   int
	err error
}

func newIOResult(n int, err error) ioResult {
	return ioResult{n: n, err: err}
}
