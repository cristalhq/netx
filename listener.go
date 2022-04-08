package netx

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"
)

// NewTCPListener returns new TCP listener for the given addr.
//
// name is used for exported metrics. Each listener in the program must have
// distinct name.
func NewTCPListener(ctx context.Context, network, addr string) (*TCPListener, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
		// ok
	default:
		return nil, fmt.Errorf("unknown network: %s", network)
	}

	ln, err := (&net.ListenConfig{}).Listen(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	tln := &TCPListener{
		Listener: ln,
		stats:    &Stats{},
	}
	return tln, err
}

// TCPListener listens for the addr passed to NewTCPListener.
//
// It also gathers various stats for the accepted connections.
type TCPListener struct {
	net.Listener
	stats *Stats
}

// Accept accepts connections from the addr passed to NewTCPListener.
func (ln *TCPListener) Accept() (net.Conn, error) {
	for {
		conn, err := ln.Listener.Accept()
		ln.stats.acceptsInc()
		if err != nil {
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			ln.stats.acceptErrorsInc()
			return nil, err
		}

		ln.stats.activeConnsInc()
		sc := &Conn{
			Conn:  conn,
			stats: ln.stats,
		}
		return sc, nil
	}
}

// Stats of the listener and accepted connections.
func (ln *TCPListener) Stats() *Stats {
	return ln.stats
}
