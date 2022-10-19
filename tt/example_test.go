package tt_test

import (
	"context"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
	"github.com/gregoryv/mq/tt/flog"
	"github.com/gregoryv/mq/tt/mux"
	"github.com/gregoryv/mq/tt/pakio"
)

func Example_runClient() {
	// replace with eg.
	// conn, _ := net.Dial("tcp", "127.0.0.1:1883")
	conn, _ := tt.Dial()

	c := tt.NewBasicClient(conn) // configure client

	fl := flog.New()
	fl.LogLevelSet(flog.LevelInfo)

	routes := []*mux.Route{
		mux.NewRoute("#", func(_ context.Context, p *mq.Publish) error {
			// handle packet...
			return nil
		}),
		mux.NewRoute("a/b"),
	}
	router := mux.NewRouter()
	router.AddRoutes(routes...)

	in := tt.NewQueue(
		[]mq.Middleware{fl.LogIncoming, fl.DumpPacket},
		func(ctx context.Context, p mq.Packet) error {
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
		},
	)
	c.InSet(in)
	c.OutStackSet([]mq.Middleware{
		fl.PrefixLoggers,
		fl.LogOutgoing,
		fl.DumpPacket,
	})

	// start handling packet flow
	ctx, _ := context.WithCancel(context.Background())
	receiver := pakio.NewReceiver(conn, in)
	go receiver.Run(ctx)

	{ // connect
		p := mq.NewConnect()
		p.SetClientID("example")
		_ = c.Send(ctx, &p)
	}
}
