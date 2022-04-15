package netx

import (
	"context"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func TestListener(t *testing.T) {
	ctx := context.Background()
	ln, err := NewTCPListener(ctx, "tcp", "127.0.0.1:8081", TCPListenerConfig{})
	failIfErr(t, err, "cannot create listener: %s", err)

	go func() {
		conn, err := ln.Accept()
		failIfErr(t, err, "cannot accept")

		_, err = conn.Write([]byte("hello world"))
		failIfErr(t, err, "cannot write")
		conn.Close()
	}()

	client, err := net.Dial("tcp", "127.0.0.1:8081")
	failIfErr(t, err, "cannot dial")

	got, err := io.ReadAll(client)
	failIfErr(t, err, "cannot read")

	if string(got) != "hello world" {
		t.Fatal(string(got))
	}
}

func TestTCPListener_DeferAccept(t *testing.T) {
	testConfig(t, TCPListenerConfig{DeferAccept: true})
}

func TestTCPListener_ReusePort(t *testing.T) {
	testConfig(t, TCPListenerConfig{ReusePort: true})
}

func TestTCPListener_FastOpen(t *testing.T) {
	testConfig(t, TCPListenerConfig{FastOpen: true})
}

func TestTCPListener_All(t *testing.T) {
	cfg := TCPListenerConfig{
		ReusePort:   true,
		DeferAccept: true,
		FastOpen:    true,
	}
	testConfig(t, cfg)
}

func TestTCPListener_Backlog(t *testing.T) {
	cfg := TCPListenerConfig{
		Backlog: 32,
	}
	testConfig(t, cfg)
}

func testConfig(t *testing.T, cfg TCPListenerConfig) {
	testTCPListener(t, cfg, "tcp4", "localhost:10081")
	// TODO(oleg): fix IPv6
	// testTCPListener(t, cfg, "tcp6", "ip6-localhost:10081")
}

func testTCPListener(t *testing.T, cfg TCPListenerConfig, network, addr string) {
	const requestsCount = 1000

	var serversCount = 1
	if cfg.ReusePort {
		serversCount = 10
	}

	doneCh := make(chan struct{}, serversCount)
	ctx := context.Background()

	var lns []net.Listener
	for i := 0; i < serversCount; i++ {
		ln, err := NewTCPListener(ctx, network, addr, cfg)
		failIfErr(t, err, "cannot create listener %d using Config %#v: %s", i, &cfg, err)

		go func() {
			serveEcho(t, ln)
			doneCh <- struct{}{}
		}()
		lns = append(lns, ln)
	}

	for i := 0; i < requestsCount; i++ {
		c, err := net.Dial(network, addr)
		failIfErr(t, err, "%d. unexpected error when dialing: %s", i, err)

		req := fmt.Sprintf("request number %d", i)
		_, err = c.Write([]byte(req))
		failIfErr(t, err, "%d. unexpected error when writing request: %s", i, err)

		err = c.(*net.TCPConn).CloseWrite()
		failIfErr(t, err, "%d. unexpected error when closing write end of the connection: %s", i, err)

		var resp []byte
		ch := make(chan struct{})
		go func() {
			resp, err = io.ReadAll(c)
			failIfErr(t, err, "%d. unexpected error when reading response: %s", i, err)
			close(ch)
		}()

		select {
		case <-ch:
		case <-time.After(200 * time.Millisecond):
			t.Fatalf("%d. timeout when waiting for response: %s", i, err)
		}

		if string(resp) != req {
			t.Fatalf("%d. unexpected response %q. Expecting %q", i, resp, req)
		}
		err = c.Close()
		failIfErr(t, err, "%d. unexpected error when closing connection: %s", i, err)
	}

	for _, ln := range lns {
		err := ln.Close()
		failIfErr(t, err, "unexpected error when closing listener")
	}

	for i := 0; i < serversCount; i++ {
		select {
		case <-doneCh:
		case <-time.After(time.Second):
			t.Fatalf("timeout when waiting for servers to be closed")
		}
	}
}

func serveEcho(t *testing.T, ln net.Listener) {
	for {
		c, err := ln.Accept()
		if err != nil {
			break
		}

		req, err := io.ReadAll(c)
		failIfErr(t, err, "unepxected error when reading request: %s", err)

		_, err = c.Write(req)
		failIfErr(t, err, "unexpected error when writing response: %s", err)

		err = c.Close()
		failIfErr(t, err, "unexpected error when closing connection: %s", err)
	}
}

func BenchmarkListener(b *testing.B) {
}

func failIfErr(tb testing.TB, err error, format string, args ...interface{}) {
	tb.Helper()
	if err != nil {
		tb.Fatalf(format, args...)
	}
}
