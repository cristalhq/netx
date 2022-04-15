package netx

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"time"
)

// TCPListenerConfig is a config TCPListener.
type TCPListenerConfig struct {
	// ReusePort enables SO_REUSEPORT.
	ReusePort bool

	// DeferAccept enables TCP_DEFER_ACCEPT.
	DeferAccept bool

	// FastOpen enables TCP_FASTOPEN.
	FastOpen bool

	// Queue length for TCP_FASTOPEN (default 256)
	FastOpenQueueLen int
}

// TCPListener listens for the addr passed to NewTCPListener.
//
// It also gathers various stats for the accepted connections.
type TCPListener struct {
	net.TCPListener
	cfg   TCPListenerConfig
	stats *Stats
}

// NewTCPListener returns new TCP listener for the given addr.
func NewTCPListener(ctx context.Context, network, addr string, cfg TCPListenerConfig) (*TCPListener, error) {
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
		cfg:         cfg,
		stats:       &Stats{},
	}
	return tln, err
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

		if err := ln.applyConfigToConn(tcpconn); err != nil {
			return nil, err
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

func (ln *TCPListener) applyConfigToConn(conn *net.TCPConn) error {
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return err
	}

	var errControl error
	rawConn.Control(func(fd uintptr) {
		errControl = ln.applyConfigToFd(int(fd))
	})
	return errControl
}

func (ln *TCPListener) applyConfigToFd(fd int) error {
	if err := disableNoDelay(fd); err != nil {
		return fmt.Errorf("cannot disable Nagle's algorithm: %s", err)
	}

	if ln.cfg.ReusePort {
		if err := enableReusePort(fd); err != nil {
			return fmt.Errorf("unable to set SO_REUSEPORT option: %s", err)
		}
	}

	if ln.cfg.FastOpen {
		queueLen := ln.cfg.FastOpenQueueLen
		if queueLen <= 0 {
			queueLen = 256
		}
		if err := enableFastOpen(fd, queueLen); err != nil {
			return fmt.Errorf("unable to set TCP_FASTOPEN option: %s", err)
		}
	}

	if ln.cfg.DeferAccept {
		if err := enableDeferAccept(fd); err != nil {
			return fmt.Errorf("unable to set TCP_DEFER_ACCEPT option: %s", err)
		}
	}

	return nil
}
