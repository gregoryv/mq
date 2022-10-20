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

	// define routing of mq.Publish packets
	complete := make(chan struct{})
	routes := []*tt.Route{
		tt.NewRoute("#", func(_ context.Context, p *mq.Publish) error {
			fmt.Println(string(p.Payload()))
			close(complete)
			return nil
		}),
	}
	router := tt.NewRouter(routes...)

	// setup outgoing queue
	pool := tt.NewIDPool(100)
	logger := tt.NewLogger(tt.LevelInfo)
	sender := tt.NewSender(conn)

	out := tt.NewQueue(
		sender.Send, // last
		logger.DumpPacket,
		logger.Out,
		pool.Out,
		logger.PrefixLoggers, //first
	)
	subscriber := tt.NewSubscriber(out, routes...)

	// we'll wait for all subscriptions to be acknowledged
	var subscribes sync.WaitGroup
	subscribes.Add(len(routes))
	waitForAllSubs := func(next mq.Handler) mq.Handler {
		return func(ctx context.Context, p mq.Packet) error {
			if _, ok := p.(*mq.SubAck); ok {
				subscribes.Done()
			}
			return next(ctx, p)
		}
	}
	// setup incoming queue
	in := tt.NewQueue(
		router.In,
		subscriber.SubscribeOnConnect,
		waitForAllSubs,
		logger.PrefixLoggers,
		pool.In,
		logger.DumpPacket,
		logger.In,
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
