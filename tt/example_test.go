package tt_test

import (
	"context"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
)

func Example_Client() {
	// replace with eg.
	// conn, _ := net.Dial("tcp", "127.0.0.1:1883")
	conn, _ := tt.Dial()

	routes := []*tt.Route{
		tt.NewRoute("#", func(_ context.Context, p *mq.Publish) error {
			// handle packet...
			return nil
		}),
		tt.NewRoute("a/b"),
	}

	router := tt.NewRouter(routes...)
	logger := tt.NewLogger(tt.LevelInfo)
	sender := tt.NewSender(conn)
	subscriber := tt.NewSubscriber(sender.Send, routes...)

	send := tt.NewQueue(
		sender.Send, // last
		logger.DumpPacket,
		logger.LogOutgoing,
		logger.PrefixLoggers, // first
	)

	in := tt.NewQueue(
		router.Route, // last
		subscriber.AutoSubscribe,
		logger.DumpPacket,
		logger.LogIncoming, // first
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
