package main

import (
	"context"
	"io"
	"net"

	"github.com/gregoryv/cmdline"
)

type Serve struct {
	Server
}

func (c *Serve) ExtraOptions(cli *cmdline.Parser) {
	c.bind = cli.Option("-b, --bind, $BIND").String("localhost:1883")
	c.acceptTimeout = cli.Option("-a, --accept-timeout").Duration("1ms")
	c.connectTimeout = cli.Option("-c, --connect-timeout").Duration("20ms")
	c.clients = make(map[string]io.ReadWriter)
}

// Run listens for tcp connections. Blocks until context is cancelled
// or accepting a connection fails. Accepting new connection can only
// be interrupted if listener has SetDeadline method.
func (c *Serve) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}
	return c.Server.Run(ln, ctx)
}
