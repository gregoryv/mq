package tt

import (
	"context"

	"github.com/gregoryv/mq"
)

func NewClient(in, out mq.Handler) *Client {
	return &Client{
		incoming: in,
		outgoing: out,
	}
}

type Client struct {
	// sequence of receivers for incoming packets
	incoming mq.Handler
	outgoing mq.Handler // first outgoing handler, set by func Run
}

// Send the packet through the outgoing idpool of handlers
func (c *Client) Send(ctx context.Context, p mq.Packet) error {
	return c.outgoing(ctx, p)
}

// Send the packet through the outgoing idpool of handlers
func (c *Client) Recv(ctx context.Context, p mq.Packet) error {
	return c.incoming(ctx, p)
}

func NewQueue(v []mq.Middleware, last mq.Handler) mq.Handler {
	if len(v) == 0 {
		return last
	}
	return v[0](NewQueue(v[1:], last))
}
