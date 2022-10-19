package tt

import (
	"context"
	"net"
	"testing"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt/intercept"
)

// thing is anything like an iot device that mostly sends stats to the
// cloud
func TestThingClient(t *testing.T) {
	recv := NewQueue([]mq.Middleware{intercept.New(0).Intercept}, mq.NoopHandler)
	send := NewQueue(nil, mq.NoopHandler)

	ctx := context.Background()

	{ // connect mq tt
		p := mq.NewConnect()
		_ = send(ctx, &p)

		ack := mq.NewConnAck()
		recv(ctx, &ack)
	}

	{ // publish application message
		p := mq.NewPublish()
		p.SetQoS(1)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		_ = send(ctx, &p)

		ack := mq.NewPubAck()
		ack.SetPacketID(p.PacketID())
		recv(ctx, &ack)
	}
	{ // disconnect nicely
		p := mq.NewDisconnect()
		if err := send(ctx, &p); err != nil {
			t.Fatal(err)
		}
	}
}

func TestClient_Send(t *testing.T) {
	_, send := NewBasicClient(&ClosedConn{})

	ctx := context.Background()
	p := mq.NewConnect()
	if err := send(ctx, &p); err == nil {
		t.Fatal("expect error")
	}
}

// ----------------------------------------

func dialBroker(t *testing.T) net.Conn {
	conn, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		t.Skip(err)
		return nil
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}
