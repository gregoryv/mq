package tt

import (
	"context"
	"testing"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt/intercept"
)

// thing is anything like an iot device that mostly sends stats to the
// cloud
func TestQueues(t *testing.T) {
	recv := NewQueue([]mq.Middleware{intercept.New(0).Intercept}, mq.NoopHandler)
	send := NewQueue(nil, mq.NoopHandler)

	ctx := context.Background()

	{ // connect mq tt
		p := mq.NewConnect()
		_ = send(ctx, &p)

		ack := mq.NewConnAck()
		recv(ctx, &ack)
	}

	{ // publish application message
		p := mq.NewPublish()
		p.SetQoS(1)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		_ = send(ctx, &p)

		ack := mq.NewPubAck()
		ack.SetPacketID(p.PacketID())
		recv(ctx, &ack)
	}
	{ // disconnect nicely
		p := mq.NewDisconnect()
		if err := send(ctx, &p); err != nil {
			t.Fatal(err)
		}
	}
}
