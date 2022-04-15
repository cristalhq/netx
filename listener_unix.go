//go:build dragonfly || linux || freebsd || netbsd

package netx

import "syscall"

func setKeepAlive(fd, secs int) error {
	if err := newError("setsockopt", syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1)); err != nil {
		return err
	}
	if err := newError("setsockopt", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, secs)); err != nil {
		return err
	}
	if err := newError("setsockopt", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPCNT, 1)); err != nil {
		return err
	}
	return newError("setsockopt", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPIDLE, secs))
}
