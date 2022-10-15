package tt_test

import (
	"context"
	"log"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
)

func init() {
	// configure logger settings before creating clients
	log.SetFlags(log.Lshortfile)
}

func Example_runClient() {

	// connect to a server, replace with eg.
	// net.Dial("tcp", "127.0.0.1:1883")
	conn, _ := tt.Dial()

	// configure client
	c := tt.NewClient()
	s := c.Settings()
	s.IOSet(conn)
	s.LogLevelSet(tt.LogLevelNone)

	router := tt.NewRouter()
	router.Add("#", func(_ context.Context, p *mq.Publish) error {
		// handle the package
		return nil
	})

	s.ReceiverSet(func(ctx context.Context, p mq.Packet) error {
		switch p.(type) {
		case *mq.ConnAck:
			// connected, maybe subscribe to topics now
			return nil
		case *mq.Publish:
			return router.Route(ctx, p)
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
