package tt

import (
	"context"
	"io"
	"io/ioutil"
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
		sender = NewSender(v)
	)

	out = NewQueue(
		sender.Out,
		logger.DumpPacket,
		logger.LogOutgoing,
		pool.SetPacketID,
	)

	in = NewQueue(
		NoopHandler,
		pool.ReusePacketID,
		logger.DumpPacket,
		logger.In,
	)

	receiver := NewReceiver(v, in)
	go receiver.Run(context.Background())

	return
}

// Dial returns a test connection to a server and the server writer
// used to send responses with.
func Dial() (*Conn, io.Writer) {
	r, w := io.Pipe()
	c := &Conn{
		Reader: r,
		Writer: ioutil.Discard, // ignore outgoing packets
	}
	return c, w
}

type Conn struct {
	io.Reader // incoming from server
	io.Writer // outgoing to server
}
