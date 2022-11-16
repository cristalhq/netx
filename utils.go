package netx

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
)

// EmptyPort looks for an empty port to listen on local interface.
func EmptyPort() (int, error) {
	for p := 30000 + rand.Intn(1000); p < 60000; p++ {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", p))
		if err == nil {
			l.Close()
			return p, nil
		}
	}
	return 0, errors.New("cannot find available port")
}

// newError same as os.NewSyscallError but shorter.
func newError(syscall string, err error) error {
	return os.NewSyscallError(syscall, err)
}
