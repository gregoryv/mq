package tt

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/gregoryv/mq"
)

var _ mq.Client = &Client{}

// thing is anything like an iot device that mostly sends stats to the
// cloud
func TestThingClient(t *testing.T) {
	c := NewClient()
	conn, server := Dial()
	c.Settings().IOSet(conn)
	ctx, incoming := runIntercepted(t, c)

	{ // connect mq tt
		p := mq.NewConnect()
		_ = c.Send(ctx, &p)

		ack := mq.NewConnAck()
		ack.WriteTo(server)

		_ = (<-incoming).(*mq.ConnAck)
	}
	{ // publish application message
		p := mq.NewPublish()
		p.SetQoS(1)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		_ = c.Send(ctx, &p)

		ack := mq.NewPubAck()
		ack.SetPacketID(p.PacketID())
		ack.WriteTo(server)
		_ = (<-incoming).(*mq.PubAck)
	}
	{ // disconnect nicely
		p := mq.NewDisconnect()
		if err := c.Send(ctx, &p); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAppClient(t *testing.T) {
	c := newClient(t)
	//c.LogLevelSet(LogLevelDebug)
	ctx, incoming := runIntercepted(t, c)

	{ // connect mq tt
		p := mq.NewConnect()
		_ = c.Send(ctx, &p)
		_ = (<-incoming).(*mq.ConnAck)
	}
	{ // subscribe
		p := mq.NewSubscribe()
		p.AddFilter("a/b", mq.OptQoS1)
		_ = c.Send(ctx, &p)
		_ = (<-incoming).(*mq.SubAck)
	}
	{ // publish application message
		p := mq.NewPublish()
		p.SetQoS(1)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		_ = c.Send(ctx, &p)
		_ = (<-incoming).(*mq.PubAck)
		_ = (<-incoming).(*mq.Publish)

	}
	{ // disconnect nicely
		p := mq.NewDisconnect()
		_ = c.Disconnect(ctx, &p)
	}
}

func TestClient_badConnect(t *testing.T) {
	c := newClient(t)
	ctx, _ := runIntercepted(t, c)
	go func() {
		if err := c.Run(ctx); err == nil {
			t.Error("Run should fail with error")
		}
	}()

	c.wire.(io.Closer).Close() // close before we write connect packet

	p := mq.NewConnect()
	if err := c.Send(ctx, &p); err == nil {
		t.Fatal("expect error")
	}
}

func TestClient_Connect_shortClientID(t *testing.T) {
	c := newClient(t)
	ctx, incoming := runIntercepted(t, c)

	p := mq.NewConnect()
	p.SetClientID("short")
	_ = c.Send(ctx, &p)
	_ = (<-incoming).(*mq.ConnAck)
}

func TestClient_Receiver(t *testing.T) {
	c := newClient(t)
	ctx, incoming := runIntercepted(t, c)

	{ // connect mq tt
		p := mq.NewConnect()
		_ = c.Send(ctx, &p)
		_ = (<-incoming).(*mq.ConnAck)
	}
}

// ----------------------------------------

func runIntercepted(t *testing.T, c *Client) (context.Context, <-chan mq.Packet) {
	r := NewInterceptor()
	c.instack = append([]mq.Middleware{r.intercept}, c.instack...) // prepend
	ctx, cancel := context.WithCancel(context.Background())
	c.Start(ctx)
	t.Cleanup(cancel)
	return ctx, r.c
}

func NewInterceptor() *Interceptor {
	return &Interceptor{
		c: make(chan mq.Packet, 0),
	}
}

type Interceptor struct {
	c chan mq.Packet
}

func (r *Interceptor) intercept(next mq.Handler) mq.Handler {
	return func(p mq.Packet) error {
		select {
		case r.c <- p: // if anyone is interested
		case <-time.After(10 * time.Millisecond):
		}
		return next(p)
	}
}

func ignore(_ mq.ControlPacket) error { return nil }

func newClient(t *testing.T) *Client {
	c := NewClient()
	s := c.Settings()
	s.IOSet(dialBroker(t))
	s.LogLevelSet(LogLevelNone)
	return c
}

func dialBroker(t *testing.T) net.Conn {
	conn, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		t.Log("no broker, did you run docker-compose up?")
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}
