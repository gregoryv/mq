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

	c := tt.NewClient()
	ctx, _ := context.WithCancel(context.Background())

	// create network connection
	conn, _ := net.Dial("tcp", "127.0.0.1:1883")

	// configure
	s := c.Settings()
	s.IOSet(conn)
	s.LogLevelSet(tt.LogLevelNone)
	s.ReceiverSet(func(p mq.Packet) error {
		switch p.(type) {
		case *mq.ConnAck:
			// connected, maybe subscribe to topics now
		}
		return nil
	})

	// start handling packet flow
	c.Start(ctx)

	// connect
	p := mq.NewConnect()
	p.SetClientID("example")
	_ = c.Connect(ctx, &p)

	// output:
}
