package tt

import (
	"context"
	"net"
	"testing"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt/intercept"
)

var _ mq.Client = &Client{}

// thing is anything like an iot device that mostly sends stats to the
// cloud
func TestThingClient(t *testing.T) {
	in := NewQueue([]mq.Middleware{intercept.New(0).Intercept}, mq.NoopHandler)
	out := NewQueue(nil, mq.NoopHandler)

	c := NewClient(in, out)

	ctx := context.Background()

	{ // connect mq tt
		p := mq.NewConnect()
		_ = c.Send(ctx, &p)

		ack := mq.NewConnAck()
		c.Recv(ctx, &ack)
	}

	{ // publish application message
		p := mq.NewPublish()
		p.SetQoS(1)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		_ = c.Send(ctx, &p)

		ack := mq.NewPubAck()
		ack.SetPacketID(p.PacketID())
		c.Recv(ctx, &ack)
	}
	{ // disconnect nicely
		p := mq.NewDisconnect()
		if err := c.Send(ctx, &p); err != nil {
			t.Fatal(err)
		}
	}
}

func TestClient_Send(t *testing.T) {
	c := NewBasicClient(&ClosedConn{})

	ctx := context.Background()
	p := mq.NewConnect()
	if err := c.Send(ctx, &p); err == nil {
		t.Fatal("expect error")
	}
}

// ----------------------------------------

func newClient(t *testing.T) *Client {
	c := NewBasicClient(dialBroker(t))
	return c
}

func dialBroker(t *testing.T) net.Conn {
	conn, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		t.Skip(err)
		return nil
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}
