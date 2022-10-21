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
	"time"

	"github.com/gregoryv/cmdline"
	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
)

func main() {
	var (
		cli    = cmdline.NewBasicParser()
		broker = cli.Option("-b --broker, $BROKER").String("127.0.0.1:1883")
		debug  = cli.Flag("-d, --debug")
	)
	cli.Parse()

	logLevel := tt.LevelInfo
	if debug {
		logLevel = tt.LevelDebug
	}

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

	// define all the features of our in/out queues
	var (
		router  = tt.NewRouter(routes...)
		logger  = tt.NewLogger(logLevel)
		sender  = tt.NewSender(conn).Out
		subwait = tt.NewSubWait(len(routes))
		conwait = tt.NewConnWait()
		pool    = tt.NewIDPool(100)

		in  = tt.NewInQueue(router.In, conwait, subwait, pool, logger)
		out = tt.NewOutQueue(sender, subwait, pool, logger)
	)

	// start handling packet flow
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

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
	<-conwait.Done()

	// suscribe all topics define by our routes
	for _, r := range routes {
		p := r.Subscribe()
		_ = out(ctx, p)
	}

	<-subwait.Done(ctx)

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
