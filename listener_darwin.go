//go:build darwin

package netx

import "syscall"

func newSocketCloexec(domain, typ, proto int) (int, error) {
	return newSocketCloexecDefault(domain, typ, proto)
}

func setKeepAlive(fd, secs int) error {
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1); err != nil {
		return err
	}

	err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, secs)
	switch err {
	case nil, syscall.ENOPROTOOPT: // OS X 10.7 and earlier don't support this option
		return syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPALIVE, secs)
	default:
		return err
	}
}
