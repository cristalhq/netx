package netx

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
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

	// Queue length for TCP_FASTOPEN (default 256).
	FastOpenQueueLen int

	// Backlog is the maximum number of pending TCP connections the listener
	// may queue before passing them to Accept.
	// Default is system-level backlog value is used.
	Backlog int
}

// TCPListener listens for the addr passed to NewTCPListener.
//
// It also gathers various stats for the accepted connections.
type TCPListener struct {
	net.Listener
	cfg   TCPListenerConfig
	stats *Stats
}

// NewTCPListener returns new TCP listener for the given addr.
func NewTCPListener(ctx context.Context, network, addr string, cfg TCPListenerConfig) (*TCPListener, error) {
	ln, err := cfg.newListener(network, addr)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	tln := &TCPListener{
		Listener: ln,
		cfg:      cfg,
		stats:    &Stats{},
	}
	return tln, err
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

func (cfg *TCPListenerConfig) newListener(network, addr string) (net.Listener, error) {
	fd, err := cfg.newSocket(network, addr)
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("netx.%d.%s.%s", os.Getpid(), network, addr)
	file := os.NewFile(uintptr(fd), name)

	ln, err := net.FileListener(file)
	if err != nil {
		file.Close()
		return nil, err
	}

	if err := file.Close(); err != nil {
		ln.Close()
		return nil, err
	}
	return ln, nil
}

func (cfg *TCPListenerConfig) newSocket(network, addr string) (fd int, err error) {
	sa, domain, err := getTCPSockaddr(network, addr)
	if err != nil {
		return 0, err
	}

	fd, err = newSocketCloexec(domain, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		return 0, err
	}

	if err := cfg.fdSetup(fd, sa, addr); err != nil {
		syscall.Close(fd)
		return 0, err
	}
	return fd, nil
}

func (cfg *TCPListenerConfig) fdSetup(fd int, sa syscall.Sockaddr, addr string) error {
	if err := newError("setsockopt", syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)); err != nil {
		return fmt.Errorf("cannot enable SO_REUSEADDR: %s", err)
	}

	// This should disable Nagle's algorithm in all accepted sockets by default.
	// Users may enable it with net.TCPConn.SetNoDelay(false).
	if err := newError("setsockopt", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1)); err != nil {
		return fmt.Errorf("cannot disable Nagle's algorithm: %s", err)
	}

	if cfg.ReusePort {
		if err := newError("setsockopt", syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, soReusePort, 1)); err != nil {
			return fmt.Errorf("cannot enable SO_REUSEPORT: %s", err)
		}
	}

	if cfg.DeferAccept {
		if err := enableDeferAccept(fd); err != nil {
			return err
		}
	}

	if cfg.FastOpen {
		if err := enableFastOpen(fd, cfg.FastOpenQueueLen); err != nil {
			return err
		}
	}

	if err := newError("bind", syscall.Bind(fd, sa)); err != nil {
		return fmt.Errorf("cannot bind to %q: %s", addr, err)
	}

	backlog := cfg.Backlog
	if backlog <= 0 {
		var err error
		if backlog, err = soMaxConn(); err != nil {
			return fmt.Errorf("cannot determine backlog to pass to listen(2): %s", err)
		}
	}
	if err := newError("listen", syscall.Listen(fd, backlog)); err != nil {
		return fmt.Errorf("cannot listen on %q: %s", addr, err)
	}

	return nil
}

func newSocketCloexecDefault(domain, typ, proto int) (int, error) {
	syscall.ForkLock.RLock()
	fd, err := syscall.Socket(domain, typ, proto)
	if err == nil {
		syscall.CloseOnExec(fd)
	}
	syscall.ForkLock.RUnlock()

	if err != nil {
		return -1, fmt.Errorf("cannot create listening socket: %s", err)
	}

	// TODO(oleg): move to fdSetup ?
	if err := newError("setnonblock", syscall.SetNonblock(fd, true)); err != nil {
		syscall.Close(fd)
		return -1, fmt.Errorf("cannot make non-blocked listening socket: %s", err)
	}
	return fd, nil
}

func getTCPSockaddr(network, addr string) (sa syscall.Sockaddr, domain int, err error) {
	tcp, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return nil, -1, err
	}

	switch network {
	case "tcp":
		return &syscall.SockaddrInet4{Port: tcp.Port}, syscall.AF_INET, nil
	case "tcp4":
		sa := &syscall.SockaddrInet4{Port: tcp.Port}
		if tcp.IP != nil {
			if len(tcp.IP) == 16 {
				copy(sa.Addr[:], tcp.IP[12:16]) // copy last 4 bytes of slice to array
			} else {
				copy(sa.Addr[:], tcp.IP) // copy all bytes of slice to array
			}
		}
		return sa, syscall.AF_INET, nil
	case "tcp6":
		sa := &syscall.SockaddrInet6{Port: tcp.Port}

		if tcp.IP != nil {
			copy(sa.Addr[:], tcp.IP) // copy all bytes of slice to array
		}

		if tcp.Zone != "" {
			iface, err := net.InterfaceByName(tcp.Zone)
			if err != nil {
				return nil, -1, err
			}

			sa.ZoneId = uint32(iface.Index)
		}
		return sa, syscall.AF_INET6, nil
	default:
		panic("unreachable")
	}
}
