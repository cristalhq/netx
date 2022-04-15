package netx

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"time"
)

// NewTCPListener returns new TCP listener for the given addr.
func NewTCPListener(ctx context.Context, network, addr string) (*TCPListener, error) {
	a, err := netip.ParseAddrPort(addr)
	if err != nil {
		return nil, err
	}
	ln, err := net.ListenTCP(network, net.TCPAddrFromAddrPort(a))
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	tln := &TCPListener{
		TCPListener: *ln,
		stats:       &Stats{},
	}
	return tln, err
}

// TCPListener listens for the addr passed to NewTCPListener.
//
// It also gathers various stats for the accepted connections.
type TCPListener struct {
	net.TCPListener
	stats *Stats
}

// Accept accepts connections from the addr passed to NewTCPListener.
func (ln *TCPListener) Accept() (net.Conn, error) {
	for {
		conn, err := ln.TCPListener.Accept()
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

		tcpconn, ok := conn.(*net.TCPConn)
		if !ok {
			panic("unreachable")
		}

		ln.stats.activeConnsInc()
		sc := &Conn{
			TCPConn: *tcpconn,
			stats:   ln.stats,
		}
		return sc, nil
	}
}

// Stats of the listener and accepted connections.
func (ln *TCPListener) Stats() *Stats {
	return ln.stats
}
