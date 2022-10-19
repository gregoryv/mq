package tt

import (
	"context"
	"testing"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt/flog"
	"github.com/gregoryv/mq/tt/idpool"
)

func BenchmarkClient_PubQoS0(b *testing.B) {
	c := NewBasicClient()
	conn, _ := Dial()
	c.IOSet(conn)
	ctx, cancel := context.WithCancel(context.Background())
	c.Start(ctx)
	defer cancel()

	for i := 0; i < b.N; i++ {
		p := mq.NewPublish()
		p.SetQoS(0)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))

		if err := c.Send(ctx, &p); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClient_PubQoS1(b *testing.B) {
	c := NewBasicClient()
	conn, server := Dial()
	c.IOSet(conn)
	ctx, cancel := context.WithCancel(context.Background())
	c.Start(ctx)
	defer cancel()

	for i := 0; i < b.N; i++ {
		p := mq.NewPublish()
		p.SetQoS(1)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))

		if err := c.Send(ctx, &p); err != nil {
			b.Fatal(err)
		}
		// response from server
		ack := mq.NewPubAck()
		ack.SetPacketID(p.PacketID())
		ack.WriteTo(server)
	}
}

// NewBasicClient returns a Queue with MaxDefaultConcurrentID and
// disabled logging
func NewBasicClient() *Queue {
	fpool := idpool.New(10)
	fl := flog.New()

	q := NewQueue()
	q.InStackSet([]mq.Middleware{
		fl.LogIncoming,
		fl.DumpPacket,
		fpool.ReusePacketID,
		fl.PrefixLoggers,
	})
	q.OutStackSet([]mq.Middleware{
		fl.PrefixLoggers,
		fpool.SetPacketID,
		fl.LogOutgoing,
		fl.DumpPacket,
	})
	return q
}
