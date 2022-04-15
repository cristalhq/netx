//go:build darwin || dragonfly || freebsd || netbsd || openbsd || rumprun
// +build darwin dragonfly freebsd netbsd openbsd rumprun

package netx

import (
	"syscall"
)

func disableNoDelay(fd int) error {
	return syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1)
}

func enableReusePort(fd int) error {
	return nil
}

func enableDeferAccept(fd int) error {
	return nil
}

func enableFastOpen(fd, _ int) error {
	return nil
}

func setBacklog(fd, backlog int) error {
	return syscall.Listen(fd, backlog)
}

func soMaxConn() (int, error) {
	return syscall.SOMAXCONN, nil
}
