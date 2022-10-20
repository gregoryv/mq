/*
Command demo tries a series of mqtt-v5 packets towards a broker, eg.
https://hub.docker.com/_/eclipse-mosquitto/

Run the broker and then

	$ go run github.com/gregoryv/mq/tt/cmd/demo
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
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/gregoryv/cmdline"
	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
)

func main() {
	var (
		cli    = cmdline.NewBasicParser()
		broker = cli.Option("-b --broker, $BROKER").String("127.0.0.1:1883")
	)
	cli.Parse()

	// connect to server
	conn, err := net.Dial("tcp", broker)
	if err != nil {
		log.Fatal(err)
	}

	// setup outgoing queue
	fpool := tt.NewIDPool(100)
	fl := tt.NewLogger()
	fl.LogLevelSet(tt.LevelInfo)
	out := tt.NewQueue(
		tt.NewSender(conn).Send, // last
		fl.DumpPacket,
		fl.LogOutgoing,
		fpool.SetPacketID,
		fl.PrefixLoggers, //first
	)

	// define routing of mq.Publish packets
	complete := make(chan struct{})
	routes := []*tt.Route{
		tt.NewRoute("#", func(_ context.Context, p *mq.Publish) error {
			close(complete)
			fmt.Println(string(p.Payload()))
			return nil
		}),
	}
	router := tt.NewRouter()
	router.AddRoutes(routes...)

	// we'll wait for all subscriptions to be acknowledged
	var subscribes sync.WaitGroup
	subscribes.Add(len(routes))

	// setup incoming queue
	in := tt.NewQueue(
		func(ctx context.Context, p mq.Packet) error {
			switch p := p.(type) {
			case *mq.ConnAck:
				// once connected we subscribe each route separately
				for _, r := range routes {
					p := mq.NewSubscribe()
					p.AddFilter(r.Filter(), 0)
					if err := out(ctx, &p); err != nil {
						log.Fatal(err)
					}
				}

			case *mq.SubAck:
				subscribes.Done()

			case *mq.Publish:
				return router.Route(ctx, p)
			}
			return nil
		},
		fl.PrefixLoggers,
		fpool.ReusePacketID,
		fl.DumpPacket,
		fl.LogIncoming,
	)

	// start handling packet flow
	ctx, _ := context.WithTimeout(context.Background(), 200*time.Millisecond)
	receiver := tt.NewReceiver(conn, in)

	go func() {
		err := receiver.Run(ctx)
		if errors.Is(err, io.EOF) {
			// client disconnected...
		}
	}()

	{ // connect
		p := mq.NewConnect()
		p.SetClientID("ttdemo")
		_ = out(ctx, &p)
	}
	subscribes.Wait()

	{ // publish
		p := mq.NewPublish()
		p.SetTopicName("a/b")
		p.SetPayload([]byte("Hello MQTT gopher friend!"))
		go out(ctx, &p)
	}

	select {
	case <-complete:
	case <-ctx.Done():
		log.Print("demo failed!")
	}
}
