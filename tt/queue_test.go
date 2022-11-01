package tt

import (
	"context"
	"testing"

	"github.com/gregoryv/mq"
)

// thing is anything like an iot device that mostly sends stats to the
// cloud
func TestQueues(t *testing.T) {
	mid := &NoopFlow{}

	out := NewOutQueue(NoopHandler, mid)
	in := NewInQueue(NoopHandler, mid, mid)

	ctx := context.Background()

	{ // connect mq tt
		_ = out(ctx, mq.NewConnect())
		in(ctx, mq.NewConnAck())
	}

	{ // publish application message
		p := mq.NewPublish()
		p.SetQoS(1)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		_ = out(ctx, p)

		ack := mq.NewPubAck()
		ack.SetPacketID(p.PacketID())
		in(ctx, ack)
	}
	{ // disconnect nicely
		p := mq.NewDisconnect()
		if err := out(ctx, p); err != nil {
			t.Fatal(err)
		}
	}
}

type NoopFlow struct{}

func (n *NoopFlow) In(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		return next(ctx, p)
	}
}

func (n *NoopFlow) Out(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		return next(ctx, p)
	}
}
