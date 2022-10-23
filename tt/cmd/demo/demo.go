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

	room := "gohpher/chat"
	pink := NewGopher("pink", broker)
	pink.Join(room)

	blue := NewGopher("blue", broker)
	blue.Join(room)

	pink.Say("hello friends")
	blue.Say("hi")

	<-time.After(10 * time.Millisecond)
}

func NewGopher(name, broker string) *Gopher {
	return &Gopher{name: name, broker: broker}
}

type Gopher struct {
	name   string
	broker string
	room   string
	net.Conn

	out mq.Handler
}

func (g *Gopher) Join(room string) {
	g.room = room
	// connect to server
	conn, err := net.Dial("tcp", g.broker)
	if err != nil {
		log.Fatal(err)
	}
	g.Conn = conn

	routes := []*tt.Route{
		tt.NewRoute(room, func(_ context.Context, p *mq.Publish) error {
			fmt.Println(string(p.Payload()))
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
	g.out = out

	// start handling packet flow
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// forward received packets to the in queue
	receiver := tt.NewReceiver(conn, in)
	go func() {
		if err := receiver.Run(ctx); errors.Is(err, io.EOF) {
			fmt.Println(g.name, "disconnected")
		}
	}()

	{ // connect
		p := mq.NewConnect()
		p.SetClientID(g.name)
		_ = out(ctx, &p)
	}
	<-conwait.Done()

	// suscribe all topics define by our routes
	for _, r := range routes {
		p := mq.NewSubscribe()
		p.AddFilter(r.Filter(), mq.OptNL)
		_ = out(ctx, &p)
	}
	<-subwait.Done(ctx)
	fmt.Println(g.name, "joined", room)
}

func (g *Gopher) Say(v string) {
	fmt.Print(g.name, "> ", v, "\n")
	p := mq.NewPublish()
	p.SetTopicName(g.room)
	p.SetPayload([]byte(g.name + ": " + v))
	g.out(context.Background(), &p)
}
