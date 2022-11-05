package main

import (
	"context"
	"net"

	"github.com/gregoryv/cmdline"
)

type ServeCmd struct {
	*Server
}

func (c *ServeCmd) ExtraOptions(cli *cmdline.Parser) {
	c.Server = NewServer()
	c.bind = cli.Option("-b, --bind, $BIND").String("localhost:1883")
	c.acceptTimeout = cli.Option("-a, --accept-timeout").Duration("1ms")
	c.connectTimeout = cli.Option("-c, --connect-timeout").Duration("20ms")
	c.poolSize = cli.Option("-p, --pool-size").Uint16(200)
}

// Run listens for tcp connections. Blocks until context is cancelled
// or accepting a connection fails. Accepting new connection can only
// be interrupted if listener has SetDeadline method.
func (c *ServeCmd) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", c.bind)
	if err != nil {
		return err
	}
	return c.Server.Run(ln, ctx)
}
