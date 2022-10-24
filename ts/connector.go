package ts

import (
	"context"

	"github.com/gregoryv/mq"
)

func NewConnector() *Connector {
	return &Connector{
		c: make(chan *mq.Connect, 0),
	}
}

type Connector struct {
	c chan *mq.Connect
}

func (c *Connector) In(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		switch p := p.(type) {
		case *mq.Connect:
			c.c <- p
		}
		return next(ctx, p)
	}
}

func (c *Connector) Done() <-chan *mq.Connect {
	return c.c
}
