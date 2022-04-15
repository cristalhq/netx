//go:build darwin

package netx

func newSocketCloexec(domain, typ, proto int) (int, error) {
	return newSocketCloexecDefault(domain, typ, proto)
}
