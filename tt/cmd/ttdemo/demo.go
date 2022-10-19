/*
Command ttdemo tries a series of mqtt-v5 packets towards a broker, eg.
https://hub.docker.com/_/eclipse-mosquitto/

Run the broker and then

	$ go run github.com/gregoryv/mq/tt/cmd/ttdemo
	ttdemo ut CONNECT ---- -------- MQTT5 ttdemo 0s 21 bytes
	ttdemo in CONNACK ---- --------  8 bytes
	ttdemo ut SUBSCRIBE --1- p1, # --r0---- 9 bytes
	ttdemo in SUBACK ---- p1 6 bytes
	ttdemo ut PUBLISH ---- p0 16 bytes
	ttdemo in PUBLISH ---- p0 16 bytes
	Hello MQTT gopher friend!
*/
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/gregoryv/cmdline"
	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
	"github.com/gregoryv/mq/tt/flog"
	"github.com/gregoryv/mq/tt/idpool"
	"github.com/gregoryv/mq/tt/mux"
)

func main() {
	var (
		cli    = cmdline.NewBasicParser()
		broker = cli.Option("-b --broker, $BROKER").String("127.0.0.1:1883")
	)
	cli.Parse()

	conn, err := net.Dial("tcp", broker)
	if err != nil {
		log.Fatal(err)
	}

	c := tt.NewQueue() // configure client

	fpool := idpool.New(100)
	fl := flog.New()
	fl.LogLevelSet(flog.LevelInfo)

	s := c.Settings()
	s.InStackSet([]mq.Middleware{
		fl.LogIncoming,
		fl.DumpPacket,
		fpool.ReusePacketID,
		fl.PrefixLoggers,
	})
	s.OutStackSet([]mq.Middleware{
		fl.PrefixLoggers,
		fpool.SetPacketID,
		fl.LogOutgoing, // keep loggers last
		fl.DumpPacket,
	})

	s.IOSet(conn)

	complete := make(chan struct{})

	routes := []*mux.Route{
		mux.NewRoute("#", func(_ context.Context, p *mq.Publish) error {
			close(complete)
			fmt.Println(string(p.Payload()))
			return nil
		}),
	}
	router := mux.NewRouter()
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
		p.SetPayload([]byte("Hello MQTT gopher friend!"))
		go c.Send(ctx, &p)
	}

	select {
	case <-complete:
	case <-ctx.Done():
		log.Print("demo failed!")
	}
}
