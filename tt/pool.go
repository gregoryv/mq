package tt

import (
	"context"

	"github.com/gregoryv/mq"
)

// Max packet id one client will use starting with 1. This also
// dictates the maximum number of packets in flight.
var MaxDefaultConcurrentID uint16 = 100

// NewPool returns a PoolFeature of reusable id's from 1..max, 0 is not used
func NewPoolFeature(max uint16) *PoolFeature {
	ids := make(chan uint16, max)
	for i := uint16(1); i <= max; i++ {
		ids <- i
	}

	return &PoolFeature{
		PoolFeature: ids,
	}
}

type PoolFeature struct {
	PoolFeature chan uint16
}

// Next returns the next available ID, blocks until one is available
// or context is canceled. Next is safe for concurrent use by multiple
// goroutines.
func (p *PoolFeature) Next(ctx context.Context) uint16 {
	select {
	case <-ctx.Done():
		return 0
	default:
		return <-p.PoolFeature
	}
}

// Reuse returns the given value to the pool
func (p *PoolFeature) Reuse(v uint16) {
	if v == 0 {
		return
	}
	p.PoolFeature <- v
}

func (o *PoolFeature) ReusePacketID(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		if p, ok := p.(mq.HasPacketID); ok {
			// todo handle dropped acks as that packet is lost. Maybe
			// a timeout for expected acks to arrive?
			if p.PacketID() > 0 {
				o.Reuse(p.PacketID())
			}
		}
		return next(ctx, p)
	}
}

// SetPacketID on outgoing packets, refs MQTT-2.2.1-3
func (o *PoolFeature) SetPacketID(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		switch p := p.(type) {
		case *mq.Publish:
			if p.QoS() > 0 {
				p.SetPacketID(o.Next(ctx))
			}

		case *mq.Subscribe:
			p.SetPacketID(o.Next(ctx))

		case *mq.Unsubscribe:
			p.SetPacketID(o.Next(ctx))
		}
		return next(ctx, p)
	}
}
