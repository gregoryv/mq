package tt_test

import (
	"context"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
	"github.com/gregoryv/mq/tt/flog"
	"github.com/gregoryv/mq/tt/mux"
)

func Example_runClient() {
	// replace with eg.
	// conn, _ := net.Dial("tcp", "127.0.0.1:1883")
	conn, _ := tt.Dial()

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

	send := tt.NewQueue(
		[]mq.Middleware{
			fl.PrefixLoggers,
			fl.LogOutgoing,
			fl.DumpPacket,
		},
		tt.NewSender(conn).Send,
	)

	in := tt.NewQueue(
		[]mq.Middleware{fl.LogIncoming, fl.DumpPacket},
		func(ctx context.Context, p mq.Packet) error {
			switch p := p.(type) {
			case *mq.ConnAck:

				// here we choose to subscribe each route separately
				for _, r := range routes {
					_ = send(ctx, r.Subscribe())
				}

			case *mq.Publish:
				return router.Route(ctx, p)
			}
			return nil
		},
	)

	// start handling packet flow
	ctx := context.Background()
	receiver := tt.NewReceiver(conn, in)
	go receiver.Run(ctx)

	{ // connect
		p := mq.NewConnect()
		p.SetClientID("example")
		_ = send(ctx, &p)
	}
}
