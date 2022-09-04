package client

import (
	"net"
	"testing"

	"github.com/gregoryv/mqtt"
)

func TestClient(t *testing.T) {
	// dial broker
	conn, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		t.Log("no broker, did you run docker-compose up?")
		t.Fatal(err)
	}

	// disconnect nicely
	defer func() {
		p := mqtt.NewDisconnect()
		t.Log(&p)
		p.WriteTo(conn)
	}()

	// connect mqtt client
	{
		p := mqtt.NewConnect()
		p.SetClientID("macy")
		p.WriteTo(conn)
		t.Log(&p)

		// check that it's acknowledged
		a, err := mqtt.ReadPacket(conn)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(a)
		if _, ok := a.(*mqtt.ConnAck); !ok {
			t.Fatal("no ConnAck")
		}
	}

	{
		p := mqtt.NewPublish()
		p.SetPacketID(99)
		p.SetRetain(true)
		p.SetTopicName("a/b/1")
		p.SetPayload([]byte("gopher"))
		p.WriteTo(conn)
		t.Log(&p)
	}
}
