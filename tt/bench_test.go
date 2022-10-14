package tt

import (
	"context"
	"testing"

	"github.com/gregoryv/mq"
)

func BenchmarkClient_PubQoS0(b *testing.B) {
	c := NewClient()
	conn, _ := Dial()
	c.Settings().IOSet(conn)
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
	c := NewClient()
	conn, server := Dial()
	c.Settings().IOSet(conn)
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
