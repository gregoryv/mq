package client

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/gregoryv/mqtt"
)

func init() {
	log.SetFlags(0)
}

func TestClient(t *testing.T) {
	// dial broker
	conn, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		t.Log("no broker, did you run docker-compose up?")
		t.Fatal(err)
	}

	c := NewClient(conn)
	// disconnect nicely
	defer func() {
		p := mqtt.NewDisconnect()
		c.Send(&p)
	}()

	// connect mqtt client
	{
		ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
		p := mqtt.NewConnect()
		if err := c.Connect(ctx, &p); err != nil {
			t.Fatal(err)
		}
		defer cancel()
	}

	// subscribe
	{
		p := mqtt.NewSubscribe()
		p.SetPacketID(101)
		p.AddFilter("a/b", mqtt.FopQoS1)
		if err := c.Subscribe(&p); err != nil {
			t.Fatal(err)
		}
	}

	// publish application message
	{
		p := mqtt.NewPublish()
		p.SetQoS(2)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		c.Publish(&p)
	}
}
