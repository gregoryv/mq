package tt

import (
	"context"
	"io"
	"net"
	"sync"
	"testing"

	"github.com/gregoryv/mq"
)

var _ mq.Client = &Client{}

// thing is anything like an iot device that mostly sends stats to the
// cloud
func TestThingClient(t *testing.T) {
	c := newClient(t)
	ctx, incoming := runIntercepted(t, c)

	{ // connect mq tt
		p := mq.NewConnect()
		_ = c.Connect(ctx, &p)
		_ = (<-incoming).(*mq.ConnAck)
	}
	{ // publish application message
		p := mq.NewPublish()
		p.SetQoS(2)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		_ = c.Pub(ctx, &p)
		_ = (<-incoming).(*mq.PubAck)
	}
	{ // disconnect nicely
		p := mq.NewDisconnect()
		if err := c.Disconnect(ctx, &p); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAppClient(t *testing.T) {
	c := newClient(t)
	ctx, incoming := runIntercepted(t, c)

	{ // connect mq tt
		p := mq.NewConnect()
		_ = c.Connect(ctx, &p)
		_ = (<-incoming).(*mq.ConnAck)
	}
	{ // subscribe
		p := mq.NewSubscribe()
		p.AddFilter("a/b", mq.FopQoS1)
		_ = c.Sub(ctx, &p)
		_ = (<-incoming).(*mq.SubAck)
	}

	// publish application message
	var wg sync.WaitGroup
	wg.Add(2) // one ack and one publish
	c.SetReceiver(func(p mq.Packet) error { wg.Done(); return nil })
	{
		p := mq.NewPublish()
		p.SetQoS(1)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		_ = c.Pub(ctx, &p)
		_ = (<-incoming).(*mq.PubAck)
		// it's not possible to do a _ = (<-incoming).(*mq.Publish) as
		// the timing is off.  so we wait for the packet to be
		// received and routed properly
		wg.Wait()
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
	if err := c.Connect(ctx, &p); err == nil {
		t.Fatal("expect error")
	}
}

func TestClient_Connect_shortClientID(t *testing.T) {
	c := newClient(t)
	ctx, incoming := runIntercepted(t, c)

	p := mq.NewConnect()
	p.SetClientID("short")
	_ = c.Connect(ctx, &p)
	_ = (<-incoming).(*mq.ConnAck)
}

func TestClient_Receiver(t *testing.T) {
	c := newClient(t)
	ctx, incoming := runIntercepted(t, c)

	v := c.Receiver()
	if v == nil {
		t.Fatal("missing initial receiver")
	}
	{ // connect mq tt
		p := mq.NewConnect()
		_ = c.Connect(ctx, &p)
		_ = (<-incoming).(*mq.ConnAck)
	}
}

// ----------------------------------------

func runIntercepted(t *testing.T, c *Client) (context.Context, chan mq.Packet) {
	ctx, cancel := context.WithCancel(context.Background())
	go c.Run(ctx)
	t.Cleanup(cancel)
	return ctx, interceptIncoming(c)
}

func interceptIncoming(c *Client) chan mq.Packet {
	ch := make(chan mq.Packet, 0)
	next := c.first
	c.first = func(p mq.Packet) error {
		select {
		case ch <- p: // if anyone is interested
		default:
		}
		return next(p)
	}
	return ch
}

func ignore(_ mq.ControlPacket) error { return nil }

func newClient(t *testing.T) *Client {
	c := NewClient()
	c.SetIO(dialBroker(t))
	c.SetLogLevel(LogLevelNone)
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
