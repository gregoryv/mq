package tt_test

import (
	"context"
	"net"
	"time"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
)

func Example_client() {
	conn, _ := net.Dial("tcp", "127.0.0.1:1883")

	routes := []*tt.Route{
		tt.NewRoute("#", func(_ context.Context, p *mq.Publish) error {
			// handle packet...
			return nil
		}),
		tt.NewRoute("a/b"),
	}

	// use middlewares and build your in/out queues with desired
	// features
	var (
		router    = tt.NewRouter(routes...)
		sender    = tt.NewSender(conn)
		subwait   = tt.NewSubWait(len(routes))
		onConnAck = make(chan *mq.ConnAck, 0)
		connwait  = tt.Intercept(onConnAck)
		pool      = tt.NewIDPool(100)
		logger    = tt.NewLogger(tt.LevelInfo)

		//                                  <-   <-    <-
		in  = tt.NewInQueue(router.In, connwait, subwait, pool, logger)
		out = tt.NewOutQueue(sender.Out, subwait, pool, logger)
	)

	// start handling packet flow
	ctx, _ := context.WithTimeout(context.Background(), 20*time.Millisecond)
	go tt.NewReceiver(in, conn).Run(ctx)

	{ // connect
		p := mq.NewConnect()
		p.SetClientID("example")
		_ = out(ctx, p)
	}
	<-onConnAck

	// connected, subscribe
	for _, r := range routes {
		p := mq.NewSubscribe()
		p.AddFilter(r.Filter(), mq.OptNL)
		_ = out(ctx, p)
	}
	<-subwait.Done(ctx)
}
