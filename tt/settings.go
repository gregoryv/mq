package tt

import (
	"io"

	"github.com/gregoryv/mq"
)

func (c *Client) InStackSet(v []mq.Middleware) error {
	if c.running {
		return ErrReadOnly
	}
	c.instack = v
	return nil
}

func (c *Client) OutStackSet(v []mq.Middleware) error {
	if c.running {
		return ErrReadOnly
	}
	c.outstack = v
	return nil
}

func (c *Client) ReceiverSet(v mq.Handler) error {
	if c.running {
		return ErrReadOnly
	}
	c.receiver = v
	return nil
}

func (c *Client) IOSet(v io.ReadWriter) error {
	if c.running {
		return ErrReadOnly
	}
	c.wire = v
	return nil
}
