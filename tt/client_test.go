package tt

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/gregoryv/mq"
)

// thing is anything like an iot device that mostly sends stats to the
// cloud
func TestThingClient(t *testing.T) {
	conn := dialBroker(t)

	c := NewNetClient(conn)
	ctx := context.Background()

	{ // connect mq tt
		p := mq.NewConnect()
		if err := c.Connect(ctx, &p); err != nil {
			t.Fatal(err)
		}
	}
	{ // publish application message
		p := mq.NewPublish()
		p.SetQoS(2)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		c.Pub(ctx, &p)
		<-time.After(50 * time.Millisecond)
	}
	{ // disconnect nicely
		p := mq.NewDisconnect()
		c.Disconnect(ctx, &p)
	}
	<-time.After(200 * time.Millisecond)
}

func TestAppClient(t *testing.T) {
	conn := dialBroker(t)

	var c mq.Client = NewNetClient(conn)
	ctx, cancel := context.WithCancel(context.Background())

	{ // connect mq tt
		p := mq.NewConnect()
		if err := c.Connect(ctx, &p); err != nil {
			t.Fatal(err)
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
		<-time.After(50 * time.Millisecond)
	}
	{ // disconnect nicely
		p := mq.NewDisconnect()
		c.Disconnect(ctx, &p)
		<-time.After(50 * time.Millisecond)
	}
	cancel()
	<-ctx.Done()
}

func TestClient_badConnect(t *testing.T) {
	conn := dialBroker(t)

	c := NewNetClient(conn)
	conn.Close() // close before we write connect packet

	p := mq.NewConnect()
	ctx := context.Background()
	if err := c.Connect(ctx, &p); err == nil {
		t.Fatal("expect error")
	}
}

func init() {
	log.SetFlags(0)
}

func ignore(_ mq.Packet) error { return nil }

func dialBroker(t *testing.T) net.Conn {
	conn, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		t.Log("no broker, did you run docker-compose up?")
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}
