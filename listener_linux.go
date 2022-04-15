//go:build linux

package netx

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
)

const (
	soReusePort = 0x0F
	tcpFastOpen = 0x17
)

func disableNoDelay(fd int) error {
	return newError("setsockopt", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, 1))
}

func enableReusePort(fd int) error {
	return newError("setsockopt", syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, soReusePort, 1))
}

func enableDeferAccept(fd int) error {
	return newError("setsockopt", syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_DEFER_ACCEPT, 1))
}

const fastOpenQueueLen = 16 * 1024

func enableFastOpen(fd int, queueLen int) error {
	return newError("setsockopt", syscall.SetsockoptInt(fd, syscall.SOL_TCP, tcpFastOpen, queueLen))
}

func soMaxConn() (int, error) {
	data, err := ioutil.ReadFile(soMaxConnFilePath)
	if err != nil {
		// This error may trigger on travis build. Just use SOMAXCONN
		if os.IsNotExist(err) {
			return syscall.SOMAXCONN, nil
		}
		return -1, err
	}
	s := strings.TrimSpace(string(data))
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return -1, fmt.Errorf("cannot parse somaxconn %q read from %s: %s", s, soMaxConnFilePath, err)
	}

	// Linux stores the backlog in a uint16.
	// Truncate number to avoid wrapping.
	// See https://github.com/golang/go/issues/5030 .
	if n > 1<<16-1 {
		n = 1<<16 - 1
	}
	return n, nil
}

const soMaxConnFilePath = "/proc/sys/net/core/somaxconn"

func setDefaultSockopts(s, family, sotype int, ipv6only bool) error {
	if family == syscall.AF_INET6 && sotype != syscall.SOCK_RAW {
		// Allow both IP versions even if the OS default
		// is otherwise. Note that some operating systems
		// never admit this option.
		err := newError("setsockopt", syscall.SetsockoptInt(s, syscall.IPPROTO_IPV6, syscall.IPV6_V6ONLY, boolint(ipv6only)))
		if err != nil {
			return err
		}
	}

	// Allow broadcast.
	return newError("setsockopt", syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1))
}

func boolint(b bool) int {
	if b {
		return 1
	}
	return 0
}
