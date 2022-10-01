package tt

import (
	"context"
	"fmt"
)

func newAckman(pool *pool) *ackman {
	return &ackman{
		pool: pool,
	}
}

// Ack manager handles a pool of packet ids that require acks.
type ackman struct {
	pool *pool
}

// Next returns next available packet id
func (a *ackman) Next(ctx context.Context) uint16 {
	return a.pool.Next(ctx)
}

func (a *ackman) Handle(ctx context.Context, packetID uint16) error {
	if !a.pool.InUse(packetID) {
		return fmt.Errorf("%v not used", packetID)
	}
	a.pool.Reuse(packetID)
	return nil
}
