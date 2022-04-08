package netx_test

import (
	"context"
	"fmt"
	"time"

	"github.com/cristalhq/netx"
)

func ExampleTCP() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ln, err := netx.NewTCPListener(ctx, "tcp", "127.0.0.1:8099")
	checkErr(err)

	stats := ln.Stats()
	fmt.Printf("total: %d\n", stats.Accepts())

	go func() {
		for {
			conn, err := ln.Accept()
			checkErr(err)

			fmt.Printf("from: %v\n", conn.LocalAddr())
		}
	}()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
