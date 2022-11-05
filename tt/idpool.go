package tt

import (
	"context"

	"github.com/gregoryv/mq"
)

// NewIDPool returns a IDPool of reusable id's from 1..max, 0 is not used
func NewIDPool(max uint16) *IDPool {
	o := IDPool{
		max:    max,
		values: make(chan uint16, max),
	}
	for i := uint16(1); i <= max; i++ {
		o.values <- i
	}
	return &o
}

type IDPool struct {
	max    uint16
	values chan uint16
}

// In checks if incoming packet has a packet ID, if so it's
// returned to the pool before next handler is called.
func (o *IDPool) In(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		switch p := p.(type) {
		case *mq.PubAck:
			// todo handle dropped acks as that packet is lost. Maybe
			// a timeout for expected acks to arrive?
			switch p.AckType() {
			case mq.PUBACK:
				o.reuse(p.PacketID())
			case mq.PUBCOMP:
				o.reuse(p.PacketID())
			}			
		}
		return next(ctx, p)
	}
}

// reuse returns the given value to the pool
func (o *IDPool) reuse(v uint16) {
	if v == 0 || v > o.max {
		return
	}
	o.values <- v
}

// Out on outgoing packets, refs MQTT-2.2.1-3
func (o *IDPool) Out(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		switch p := p.(type) {
		case *mq.Publish:
			// Don't set packet id if already set, this is used in
			// eg. PubRel packets
			if p.QoS() > 0 && p.PacketID() == 0 {
				p.SetPacketID(o.next(ctx))
			}

		case *mq.Subscribe:
			p.SetPacketID(o.next(ctx))

		case *mq.Unsubscribe:
			p.SetPacketID(o.next(ctx))
		}
		return next(ctx, p)
	}
}

// next returns the next available ID, blocks until one is available
// or context is canceled. next is safe for concurrent use by multiple
// goroutines.
func (o *IDPool) next(ctx context.Context) uint16 {
	select {
	case <-ctx.Done():
	case v := <-o.values:
		return v
	}
	return 0
}
