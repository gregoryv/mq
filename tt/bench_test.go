package tt

import (
	"context"
	"io"
	"testing"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt/flog"
	"github.com/gregoryv/mq/tt/idpool"
)

func BenchmarkClient_PubQoS0(b *testing.B) {
	conn, _ := Dial()
	c := NewBasicClient(conn)
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
	conn, server := Dial()
	c := NewBasicClient(conn)

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

// NewBasicClient returns a Client with MaxDefaultConcurrentID and
// disabled logging
func NewBasicClient(v io.ReadWriter) *Client {
	fpool := idpool.New(10)
	fl := flog.New()

	in := NewQueue([]mq.Middleware{
		fl.LogIncoming,
		fl.DumpPacket,
		fpool.ReusePacketID,
		fl.PrefixLoggers,
	}, mq.NoopHandler)

	c := NewClient()
	c.IOSet(v)
	c.InSet(in)

	c.OutStackSet([]mq.Middleware{
		fl.PrefixLoggers,
		fpool.SetPacketID,
		fl.LogOutgoing,
		fl.DumpPacket,
	})
	return c
}
