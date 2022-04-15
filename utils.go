package netx

import "os"

// newError same as os.NewSyscallError but shorter.
func newError(syscall string, err error) error {
	return os.NewSyscallError(syscall, err)
}
