//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package netx

import (
	"syscall"
)

const soReusePort = syscall.SO_REUSEPORT

func disableNoDelay(fd int) error {
	return nil
}

func enableReusePort(fd int) error {
	return nil
}

func enableDeferAccept(fd int) error {
	return nil
}

func enableFastOpen(fd, queueLen int) error {
	return nil
}

func setBacklog(fd, backlog int) error {
	return syscall.Listen(fd, backlog)
}

func soMaxConn() (int, error) {
	return syscall.SOMAXCONN, nil
}
