/*
Command demo tries a series of mqtt-v5 packets towards a broker, eg.
https://hub.docker.com/_/eclipse-mosquitto/

Run the broker and then

	$ go run github.com/gregoryv/mq/tt/cmd/demo
	pink joined gohpher/chat
	blue joined gohpher/chat
	pink> hello friends
	blue> hi
	pink: hello friends
	blue: hi
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

var logLevel = tt.LevelNone

func main() {
	var (
		cli    = cmdline.NewBasicParser()
		broker = cli.Option("-b --broker, $BROKER").String("127.0.0.1:1883")
		debug  = cli.Flag("-d, --debug")
	)
	cli.Parse()

	if debug {
		logLevel = tt.LevelInfo
	}

	room := "gophers/chat"
	pink := NewGopher("pink")
	pink.Join(broker, room)

	blue := NewGopher("blue")
	blue.Join(broker, room)

	pink.Say("hello friends")
	blue.Say("hi")

	<-time.After(10 * time.Millisecond)
}

func NewGopher(name string) *Gopher {
	return &Gopher{
		Name: name,
		Say:  func(string) {}, // noop
	}
}

type Gopher struct {
	Name string
	Say  func(v string)
}

func (g *Gopher) Join(broker, room string) {
	// connect to server
	conn, err := net.Dial("tcp", broker)
	if err != nil {
		log.Fatal(err)
	}

	routes := []*tt.Route{
		tt.NewRoute(room, func(_ context.Context, p *mq.Publish) error {
			fmt.Println(string(p.Payload()))
			return nil
		}),
	}

	// define all the features of our in/out queues
	var (
		router    = tt.NewRouter(routes...)
		logger    = tt.NewLogger(logLevel)
		sender    = tt.NewSender(conn).Out
		subwait   = tt.NewSubWait(len(routes))
		onConnAck = make(chan *mq.ConnAck, 0)
		conwait   = tt.Intercept(onConnAck)
		pool      = tt.NewIDPool(100)

		in  = tt.NewInQueue(router.In, conwait, subwait, pool, logger)
		out = tt.NewOutQueue(sender, logger, subwait, pool)
	)

	// start handling packet flow
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// forward received packets to the in queue
	receiver := tt.NewReceiver(in, conn)
	go func() {
		if err := receiver.Run(ctx); errors.Is(err, io.EOF) {
			fmt.Println(g.Name, "disconnected")
		}
	}()

	{ // connect
		p := mq.NewConnect()
		p.SetClientID(g.Name)
		_ = out(ctx, p)
	}
	<-onConnAck

	// suscribe all topics define by our routes
	for _, r := range routes {
		p := mq.NewSubscribe()
		p.AddFilter(r.Filter(), mq.OptNL)
		_ = out(ctx, p)
	}
	<-subwait.Done(ctx)
	fmt.Println(g.Name, "joined", room)

	g.Say = func(v string) {
		fmt.Print(g.Name, "> ", v, "\n")
		p := mq.NewPublish()
		p.SetTopicName(room)
		p.SetPayload([]byte(g.Name + ": " + v))
		out(context.Background(), p)
	}
}
