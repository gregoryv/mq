package tt_test

import (
	"context"
	"time"

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

	var (
		router     = tt.NewRouter(routes...)
		logger     = tt.NewLogger(tt.LevelInfo)
		sender     = tt.NewSender(conn)
		subscriber = tt.NewSubscriber(sender.Out, routes...)
		ackwait    = tt.NewSubWait(len(routes))
	)

	send := tt.NewQueue(
		sender.Out, // last

		ackwait.Out,

		logger.Out,
	)

	in := tt.NewQueue(
		router.In, // last

		ackwait.In,
		subscriber.SubscribeOnConnect,

		logger.In,
	)

	// start handling packet flow
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Millisecond)
	receiver := tt.NewReceiver(conn, in)
	go receiver.Run(ctx)

	{ // connect
		p := mq.NewConnect()
		p.SetClientID("example")
		_ = send(ctx, &p)
	}
	<-ackwait.AllSubscribed(ctx)
}
