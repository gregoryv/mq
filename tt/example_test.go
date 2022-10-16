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
	// replace with eg.
	// conn, _ := net.Dial("tcp", "127.0.0.1:1883")
	conn, _ := tt.Dial()

	c := tt.NewBasicClient() // configure client
	s := c.Settings()
	s.IOSet(conn)

	routes := []*tt.Route{
		tt.NewRoute("#", func(_ context.Context, p *mq.Publish) error {
			// handle packet...
			return nil
		}),
		tt.NewRoute("a/b"),
	}
	router := tt.NewRouter()
	router.AddRoutes(routes...)

	s.ReceiverSet(func(ctx context.Context, p mq.Packet) error {
		switch p := p.(type) {
		case *mq.ConnAck:

			// here we choose to subscribe each route separately
			for _, r := range routes {
				_ = c.Send(ctx, r.Subscribe())
			}

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
}
