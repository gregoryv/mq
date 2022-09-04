package client

import (
	"log"
	"net"
	"testing"

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
		p := mqtt.NewConnect()
		p.SetClientID("macy")
		if err := c.Connect(&p); err != nil {
			t.Fatal(err)
		}
	}

	// publish application message
	{
		p := mqtt.NewPublish()
		p.SetPacketID(99)
		p.SetRetain(true)
		p.SetTopicName("a/b/1")
		p.SetPayload([]byte("gopher"))
		c.Send(&p)
	}
}
