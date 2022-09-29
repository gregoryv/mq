package tt

import (
	"context"
	"fmt"
)

func NewAckman(pool *IDPool) *Ackman {
	return &Ackman{
		pool: pool,
	}
}

// Ack manager handles a pool of packet ids that require acks.
type Ackman struct {
	pool *IDPool
}

// Next returns next available packet id
func (a *Ackman) Next(ctx context.Context) uint16 {
	return a.pool.Next(ctx)
}

func (a *Ackman) Handle(ctx context.Context, ack PubSubAck) error {
	if v := ack.PacketID(); !a.pool.InUse(v) {
		return fmt.Errorf("%v not used", v)
	}
	a.pool.Reuse(ack.PacketID())
	return nil
}

type PubSubAck interface {
	PacketID() uint16
}
