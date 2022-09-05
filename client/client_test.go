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
	ctx, cancel := context.WithCancel(context.Background())

	// connect mqtt client
	{
		p := mqtt.NewConnect()
		if err := c.Connect(ctx, &p); err != nil {
			t.Fatal(err)
		}
	}

	// subscribe
	{
		p := mqtt.NewSubscribe()
		p.SetPacketID(101)
		p.AddFilter("a/b", mqtt.FopQoS1|mqtt.FopNL)
		if err := c.Subscribe(&p); err != nil {
			t.Fatal(err)
		}
		<-time.After(50 * time.Millisecond)
	}
	// publish application message
	{
		p := mqtt.NewPublish()
		// p.SetQoS(1) seems we get a malformed error with this
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		c.Publish(&p)
		<-time.After(50 * time.Millisecond)
	}

	// disconnect nicely
	{
		p := mqtt.NewDisconnect()
		c.Disconnect(&p)
	}
	<-time.After(200 * time.Millisecond)
	cancel()
	<-ctx.Done()

}
