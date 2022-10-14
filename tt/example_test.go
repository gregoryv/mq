package tt_test

import (
	"context"
	"log"
	"net"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
)

func init() {
	// configure logger settings before creating clients
	log.SetFlags(log.Lshortfile)
}

func Example_runClient() {
	c := tt.NewClient()

	// configure
	s := c.Settings()

	// create network connection
	conn, _ := net.Dial("tcp", "127.0.0.1:1883")
	s.IOSet(conn)
	s.LogLevelSet(tt.LogLevelNone)
	s.ReceiverSet(func(_ context.Context, p mq.Packet) error {
		switch p.(type) {
		case *mq.ConnAck:
			// connected, maybe subscribe to topics now
		}
		return nil
	})

	// start handling packet flow
	ctx, _ := context.WithCancel(context.Background())
	c.Start(ctx)

	{ // connect
		p := mq.NewConnect()
		p.SetClientID("example")
		_ = c.Send(ctx, &p)
	}
	{ // subscribe
		p := mq.NewSubscribe()
		p.AddFilter("a/b", mq.OptQoS1)
		_ = c.Send(ctx, &p)
	}
}
