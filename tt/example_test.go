package tt_test

import (
	"context"
	"log"
	"net"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

func Example_runClient() {
	// create network connection
	conn, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		log.Fatal(err)
	}

	// configure
	c := tt.NewClient()
	c.IOSet(conn)
	c.ReceiverSet(func(p mq.Packet) error {
		// do something with it ...
		// todo specify when errors should be returned by receivers
		return nil
	})

	// start handling packet flow
	ctx, _ := context.WithCancel(context.Background())
	go c.Run(ctx) //todo how to handle that last error

	// output:
}

func ExampleClient_Connect() {
	c := tt.NewClient()
	ctx, _ := context.WithCancel(context.Background())

	// create network connection
	conn, _ := net.Dial("tcp", "127.0.0.1:1883")

	// configure
	c.IOSet(conn)
	c.LogLevelSet(tt.LogLevelNone)
	c.ReceiverSet(func(p mq.Packet) error {
		switch p.(type) {
		case *mq.ConnAck:
			// connected, maybe subscribe to topics now
		}
		return nil
	})

	// start handling packet flow
	go c.Run(ctx)

	// connect
	p := mq.NewConnect()
	p.SetClientID("example")
	_ = c.Connect(ctx, &p)

	// output:
}

// todo maybe add a mechanism for sequenced packets
