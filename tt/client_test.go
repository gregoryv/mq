package tt

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/gregoryv/mq"
)

var _ mq.Client = &Client{}

// thing is anything like an iot device that mostly sends stats to the
// cloud
func TestThingClient(t *testing.T) {
	conn := dialBroker(t)

	c := NewNetClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	go c.Run(ctx)
	t.Cleanup(cancel)
	incoming := interceptIncoming(c)

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
	conn := dialBroker(t)

	c := NewNetClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	go c.Run(ctx)
	t.Cleanup(cancel)
	incoming := interceptIncoming(c)

	{ // connect mq tt
		p := mq.NewConnect()
		_ = c.Connect(ctx, &p)
		if p, ok := (<-incoming).(*mq.ConnAck); !ok {
			t.Error("expected ack, got", p)
		}
	}
	{ // subscribe
		p := mq.NewSubscribe()
		p.AddFilter("a/b", mq.FopQoS1)
		_ = c.Sub(ctx, &p)
		_ = (<-incoming).(*mq.SubAck)
	}
	// todo use a client to publish an application message on the
	// subscribed topic wip, need to implement routing of subscribed
	// filters in previous step and assert that the message arrives
	// properly.
	{ // publish application message
		p := mq.NewPublish()
		p.SetQoS(1)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		_ = c.Pub(ctx, &p)
		_ = (<-incoming).(*mq.PubAck)
	}
	{ // disconnect nicely
		p := mq.NewDisconnect()
		_ = c.Disconnect(ctx, &p)
	}
}

func TestClient_badConnect(t *testing.T) {
	conn := dialBroker(t)

	c := NewNetClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		if err := c.Run(ctx); err == nil {
			t.Error("Run should fail with error")
		}
	}()

	conn.Close() // close before we write connect packet

	p := mq.NewConnect()
	if err := c.Connect(ctx, &p); err == nil {
		t.Fatal("expect error")
	}
}

func init() {
	log.SetFlags(0)
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

func dialBroker(t *testing.T) net.Conn {
	conn, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		t.Log("no broker, did you run docker-compose up?")
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}
