package tt

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

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

	{ // connect mq tt
		p := mq.NewConnect()
		_ = c.Connect(ctx, &p)
		if p, ok := (<-c.Incoming).(*mq.ConnAck); !ok {
			t.Error("expected ack, got", p)
		}
	}
	{ // publish application message
		p := mq.NewPublish()
		p.SetQoS(2)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		c.Pub(ctx, &p)

		if p, ok := (<-c.Incoming).(*mq.PubAck); !ok {
			t.Error("expected ack, got", p)
		}
	}
	{ // disconnect nicely
		p := mq.NewDisconnect()
		if err := c.Disconnect(ctx, &p); err != nil {
			t.Fatal(err)
		}
	}
	<-time.After(200 * time.Millisecond)
}

func TestAppClient(t *testing.T) {
	conn := dialBroker(t)

	c := NewNetClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	go c.Run(ctx)
	t.Cleanup(cancel)

	{ // connect mq tt
		p := mq.NewConnect()
		_ = c.Connect(ctx, &p)
		if p, ok := (<-c.Incoming).(*mq.ConnAck); !ok {
			t.Error("expected ack, got", p)
		}
	}
	{ // subscribe
		p := mq.NewSubscribe()
		p.AddFilter("a/b", mq.FopQoS1)
		if err := c.Sub(ctx, &p); err != nil {
			t.Fatal(err)
		}
	}
	// todo use a client to publish an application message on the
	// subscribed topic wip, need to implement routing of subscribed
	// filters in previous step and assert that the message arrives
	// properly.
	{
		// publish application message
		p := mq.NewPublish()
		p.SetQoS(2)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		c.Pub(ctx, &p)
		if p, ok := (<-c.Incoming).(*mq.PubAck); !ok {
			t.Error("expected ack, got", p)
		}
	}
	{ // disconnect nicely
		p := mq.NewDisconnect()
		c.Disconnect(ctx, &p)
		<-time.After(50 * time.Millisecond)
	}
}

func TestClient_badConnect(t *testing.T) {
	conn := dialBroker(t)

	c := NewNetClient(conn)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		if err := c.Run(ctx); err == nil {
			t.Error(err)
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
