package tt

import (
	"context"
	"io"
	"testing"

	"github.com/gregoryv/mq"
)

func BenchmarkClient_PubQoS0(b *testing.B) {
	conn, _ := Dial()
	send, _ := NewClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < b.N; i++ {
		p := mq.NewPublish()
		p.SetQoS(0)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))

		if err := send(ctx, &p); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClient_PubQoS1(b *testing.B) {
	conn, server := Dial()
	send, _ := NewClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < b.N; i++ {
		p := mq.NewPublish()
		p.SetQoS(1)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))

		if err := send(ctx, &p); err != nil {
			b.Fatal(err)
		}
		// response from server
		ack := mq.NewPubAck()
		ack.SetPacketID(p.PacketID())
		ack.WriteTo(server)
	}
}

// NewClient returns a Client
func NewClient(v io.ReadWriter) (out mq.Handler, in mq.Handler) {
	var (
		pool   = NewIDPool(10)
		logger = NewLogger(LevelNone)
		sender = NewSender(v).Out
	)
	out = NewOutQueue(sender, logger, pool)
	in = NewInQueue(NoopHandler, pool, logger)
	go NewReceiver(v, in).Run(context.Background())
	return
}
