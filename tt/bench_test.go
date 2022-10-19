package tt

import (
	"context"
	"io"
	"io/ioutil"
	"testing"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt/flog"
	"github.com/gregoryv/mq/tt/idpool"
	"github.com/gregoryv/mq/tt/pakio"
)

func BenchmarkClient_PubQoS0(b *testing.B) {
	conn, _ := Dial()
	_, send := NewBasicClient(conn)
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
	_, send := NewBasicClient(conn)

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

// NewBasicClient returns a Client with MaxDefaultConcurrentID and
// disabled logging
func NewBasicClient(v io.ReadWriter) (in mq.Handler, out mq.Handler) {
	fpool := idpool.New(10)
	fl := flog.New()

	in = NewQueue([]mq.Middleware{
		fl.LogIncoming,
		fl.DumpPacket,
		fpool.ReusePacketID,
		fl.PrefixLoggers,
	}, mq.NoopHandler)

	receiver := pakio.NewReceiver(v, in)
	go receiver.Run(context.Background())

	out = NewQueue(
		[]mq.Middleware{
			fl.PrefixLoggers,
			fpool.SetPacketID,
			fl.LogOutgoing,
			fl.DumpPacket,
		},
		pakio.NewSender(v).Send,
	)

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
