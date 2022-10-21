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
	conn, server := tt.Dial()

	routes := []*tt.Route{
		tt.NewRoute("#", func(_ context.Context, p *mq.Publish) error {
			// handle packet...
			return nil
		}),
		tt.NewRoute("a/b"),
	}

	var (
		router  = tt.NewRouter(routes...)
		logger  = tt.NewLogger(tt.LevelInfo)
		sender  = tt.NewSender(conn).Out
		subwait = tt.NewSubWait(len(routes))
		conwait = tt.NewConnWait()
		pool = tt.NewIDPool(100)

		out = tt.NewOutQueue(sender, subwait, pool, logger)
		in  = tt.NewInQueue(router.In, conwait, subwait, pool, logger)
	)

	// start handling packet flow
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Millisecond)
	receiver := tt.NewReceiver(conn, in)
	go receiver.Run(ctx)

	{ // connect
		p := mq.NewConnect()
		p.SetClientID("example")
		_ = out(ctx, &p)
		server.Ack(&p) // mock server response
	}
	<-conwait.Done()

	for _, r := range routes {
		p := r.Subscribe()
		_ = out(ctx, p)
		server.Ack(p)
	}

	<-subwait.AllSubscribed(ctx)
	// output:
}
