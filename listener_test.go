package netx

import (
	"context"
	"io"
	"net"
	"testing"
)

func TestListener(t *testing.T) {
	ctx := context.Background()
	ln, err := NewTCPListener(ctx, "tcp", "127.0.0.1:8081", TCPListenerConfig{})
	failIfErr(t, err)

	go func() {
		conn, err := ln.Accept()
		failIfErr(t, err)

		_, err = conn.Write([]byte("hello world"))
		failIfErr(t, err)
		conn.Close()
	}()

	client, err := net.Dial("tcp", "127.0.0.1:8081")
	failIfErr(t, err)

	got, err := io.ReadAll(client)
	failIfErr(t, err)

	if string(got) != "hello world" {
		t.Fatal(string(got))
	}
}

func BenchmarkListener(b *testing.B) {
}

func failIfErr(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatal(err)
	}
}
