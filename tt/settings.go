package tt

import (
	"fmt"

	"github.com/gregoryv/mq"
)

type Settings interface {
	// ReceiverSet configures receiver for any incoming mq.Publish
	// packets. The client handles PacketID reuse.
	ReceiverSet(mq.Handler) error
}

type setWrite struct {
	setRead
}

func (s *setWrite) ReceiverSet(v mq.Handler) error {
	s.c.receiver = v
	return nil
}

type setRead struct {
	c *Client
}

func (s *setRead) ReceiverSet(_ mq.Handler) error { return ErrReadOnly }

var ErrReadOnly = fmt.Errorf("read only")
