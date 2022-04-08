package netx

import (
	"errors"
	"io"
	"net"
	"sync"
)

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
