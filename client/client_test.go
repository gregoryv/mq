package client

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/gregoryv/mqtt"
)

// thing is anything like an iot device that mostly sends stats to the
// cloud
func TestThingClient(t *testing.T) {
	//dial broker
	conn, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		t.Log("no broker, did you run docker-compose up?")
		t.Fatal(err)
	}

	c := NewNetClient(conn)
	ctx, cancel := context.WithCancel(context.Background())

	{ // connect mqtt client
		p := mqtt.NewConnect()
		if err := c.Connect(ctx, &p); err != nil {
			t.Fatal(err)
		}
	}
	{ // publish application message
		p := mqtt.NewPublish()
		p.SetQoS(2)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		c.Publish(ctx, &p)
		<-time.After(50 * time.Millisecond)
	}
	{ // disconnect nicely
		p := mqtt.NewDisconnect()
		c.Disconnect(&p)
	}
	<-time.After(200 * time.Millisecond)
	cancel()
	<-ctx.Done()
}

func TestAppClient(t *testing.T) {
	// dial broker
	conn, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		t.Log("no broker, did you run docker-compose up?")
		t.Fatal(err)
	}

	c := NewNetClient(conn)
	ctx, cancel := context.WithCancel(context.Background())

	{ // connect mqtt client
		p := mqtt.NewConnect()
		if err := c.Connect(ctx, &p); err != nil {
			t.Fatal(err)
		}
	}
	{ // subscribe
		p := mqtt.NewSubscribe()
		p.AddFilter("a/b", mqtt.FopQoS1)
		if err := c.Subscribe(ctx, &p); err != nil {
			t.Fatal(err)
		}
	}
	// todo use a client to send a message on the subscribed topic
	// wip, need to implement routing of subscribed filters in
	// previous step and assert that the message arrives properly.
	{
		// publish application message
		p := mqtt.NewPublish()
		p.SetQoS(2)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		c.Publish(ctx, &p)
		<-time.After(50 * time.Millisecond)
	}
	{ // disconnect nicely
		p := mqtt.NewDisconnect()
		c.Disconnect(&p)
		<-time.After(50 * time.Millisecond)
	}
	cancel()
	<-ctx.Done()
}

func TestClient_badConnect(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		t.Log("no broker, did you run docker-compose up?")
		t.Fatal(err)
	}

	c := NewNetClient(conn)
	conn.Close()

	p := mqtt.NewConnect()
	ctx := context.Background()
	if err := c.Connect(ctx, &p); err == nil {
		t.Fatal("expect error")
	}
}

func init() {
	log.SetFlags(0)
}
