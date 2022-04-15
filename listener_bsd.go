//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package netx

import (
	"runtime"
	"syscall"
)

const soReusePort = syscall.SO_REUSEPORT

func disableNoDelay(fd int) error {
	return newError("setsockopt", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1))
}

func enableReusePort(fd int) error {
	return newError("setsockopt", syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1))
}

func enableDeferAccept(fd int) error {
	return nil
}

func enableFastOpen(fd, queueLen int) error {
	return nil
}

func setBacklog(fd, backlog int) error {
	return newError("listen", syscall.Listen(fd, backlog))
}

func setLinger(fd, sec int) error {
	var l syscall.Linger
	if sec >= 0 {
		l.Onoff, l.Linger = 1, int32(sec)
	}
	return newError("setsockopt", syscall.SetsockoptLinger(fd, syscall.SOL_SOCKET, syscall.SO_LINGER, &l))
}

func soMaxConn() (int, error) {
	return syscall.SOMAXCONN, nil
}

func setDefaultSockopts(fd, family, sotype int, ipv6only bool) error {
	if runtime.GOOS == "dragonfly" && sotype != syscall.SOCK_RAW {
		// On DragonFly BSD, we adjust the ephemeral port range because unlike other BSD systems its default
		// port range doesn't conform to IANA recommendation as described in RFC 6056 and is pretty narrow.
		var err error
		switch family {
		case syscall.AF_INET:
			err = newError("setsockopt", syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_PORTRANGE, syscall.IP_PORTRANGE_HIGH))
		case syscall.AF_INET6:
			err = newError("setsockopt", syscall.SetsockoptInt(fd, syscall.IPPROTO_IPV6, syscall.IPV6_PORTRANGE, syscall.IPV6_PORTRANGE_HIGH))
		}
		if err != nil {
			return err
		}
	}
	return newError("setsockopt", syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1))
}
