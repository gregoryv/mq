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

func init() {
	// configure logger settings before creating clients
	log.SetFlags(0)
}

func main() {
	conn, _ := net.Dial("tcp", "127.0.0.1:1883")

	c := tt.NewClient() // configure client
	s := c.Settings()
	s.IOSet(conn)
	s.LogLevelSet(tt.LogLevelInfo)

	complete := make(chan struct{})

	router := tt.NewRouter()
	router.Add("#", func(_ context.Context, p *mq.Publish) error {
		complete <- struct{}{}
		return nil
	})

	var subscribes sync.WaitGroup
	subscribes.Add(len(router.Routes()))
	s.ReceiverSet(func(ctx context.Context, p mq.Packet) error {
		switch p := p.(type) {
		case *mq.ConnAck:
			// here we choose to subscribe each route separately
			for _, r := range router.Routes() {
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
			{
				ack := mq.NewPubAck()
				ack.SetPacketID(p.PacketID())
				_ = c.Send(ctx, &ack)
			}
			return router.Route(ctx, p)
		}
		return nil
	})

	// start handling packet flow
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	c.Start(ctx)

	{ // connect
		p := mq.NewConnect()
		p.SetClientID("ttdemo")
		_ = c.Send(ctx, &p)
	}

	subscribes.Wait()
	{ // publish
		p := mq.NewPublish()
		p.SetQoS(1)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		_ = c.Send(ctx, &p)
	}

	select {
	case <-complete:
		log.Print("complete!")
	case <-ctx.Done():
		log.Print("failed!")
	}

}
