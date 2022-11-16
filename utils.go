package netx

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
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

// WaitPort until you port is free.
func WaitPort(ctx context.Context, port int) error {
	return WaitAddr(ctx, fmt.Sprintf(":%d", port))
}

// WaitAddr until you addr is free.
func WaitAddr(ctx context.Context, addr string) error {
	errCh := make(chan error, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				errCh <- fmt.Errorf("cannot connect to %s", addr)
				return
			default:
				c, err := net.Dial("tcp", addr)
				if err == nil {
					c.Close()
					errCh <- nil
					return
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

// newError same as os.NewSyscallError but shorter.
func newError(syscall string, err error) error {
	return os.NewSyscallError(syscall, err)
}
