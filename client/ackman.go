package client

import "context"

func NewAckman(pool *IDPool) *Ackman {
	return &Ackman{
		pool: pool,
	}
}

type Ackman struct {
	pool *IDPool
}

func (a *Ackman) Next(ctx context.Context, wait bool) uint16 {
	if wait == true {
		panic("todo implement Ackman.Next(_, true)")
	}
	return a.pool.Next(ctx)
}

func (a *Ackman) Handle(ctx context.Context, ack PubSubAck) {
	a.pool.Reuse(ack.PacketID())
}

type PubSubAck interface {
	PacketID() uint16
}
