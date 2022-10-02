package tt_test

import (
	"bytes"
	"context"
	"log"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
)

func Example_newClient() {
	c := tt.NewClient()

	// configure
	var conn bytes.Buffer // replace with e.g. net.Conn
	c.SetIO(&conn)
	c.SetReceiver(func(p mq.Packet) error {

		switch p := p.(type) {
		case *mq.Publish:
			// handle incoming publish packet

		case *mq.PubAck: // includes PubRec, PubRel PubComp
			switch p.AckType() {
			case mq.PUBACK:
			case mq.PUBREC:
			case mq.PUBCOMP:
			case mq.PUBREL:
			}
		case *mq.SubAck:
		case *mq.ConnAck:
			_ = p
		}
		log.Print(p)
		// todo specify when errors should be returned by receivers
		return nil
	})

	// start handling packet flow
	ctx, _ := context.WithCancel(context.Background())
	go c.Run(ctx)

	// connect
	cp := mq.NewConnect()
	cp.SetClientID("gopher")
	_ = c.Connect(ctx, &cp)

	// subscribe
	sp := mq.NewSubscribe()
	sp.AddFilter("a/b", mq.FopQoS1)
	_ = c.Sub(ctx, &sp)

	// publish
	pp := mq.NewPublish()
	pp.SetQoS(1)
	pp.SetTopicName("a/b")
	pp.SetPayload([]byte("gopher"))
	_ = c.Pub(ctx, &pp) // todo if Pub and Sub only can fail on
	// send errors then c.Run will fail so
	// there is no reason for returning error here?

	// disconnect
	dp := mq.NewDisconnect()
	_ = c.Disconnect(ctx, &dp)

}
