package main

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
)

func main() {
	conn, _ := net.Dial("tcp", "127.0.0.1:1883")

	c := tt.NewBasicClient() // configure client
	s := c.Settings()
	s.IOSet(conn)
	s.LogLevelSet(tt.LogLevelDebug)

	complete := make(chan struct{})

	routes := []*tt.Route{
		tt.NewRoute("#", func(_ context.Context, p *mq.Publish) error {
			close(complete)
			return nil
		}),
	}
	router := tt.NewRouter()
	router.AddRoutes(routes...)

	var subscribes sync.WaitGroup
	subscribes.Add(len(routes))

	s.ReceiverSet(func(ctx context.Context, p mq.Packet) error {
		switch p := p.(type) {
		case *mq.ConnAck:
			// here we choose to subscribe each route separately
			for _, r := range routes {
				{
					p := mq.NewSubscribe()
					p.AddFilter(r.Filter(), 0)
					if err := c.Send(ctx, &p); err != nil {
						log.Fatal(err)
					}
				}
			}

		case *mq.SubAck:
			subscribes.Done()

		case *mq.Publish:
			return router.Route(ctx, p)
		}
		return nil
	})

	// start handling packet flow
	ctx, _ := context.WithTimeout(context.Background(), 200*time.Millisecond)
	c.Start(ctx)

	{ // connect
		p := mq.NewConnect()
		p.SetClientID("ttdemo")
		_ = c.Send(ctx, &p)
	}

	subscribes.Wait()
	{ // publish
		p := mq.NewPublish()
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		go c.Send(ctx, &p)
	}

	select {
	case <-complete:
		log.Print("demo complete!")
	case <-ctx.Done():
		log.Print("demo failed!")
	}

}
