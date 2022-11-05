package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"time"

	"github.com/gregoryv/cmdline"
	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
)

type Pub struct {
	server *url.URL

	topic   string
	payload string
	qos     uint16
}

func (c *Pub) ExtraOptions(cli *cmdline.Parser) {
	c.server = cli.Option("-s, --server").Url("localhost:1883")
	c.topic = cli.Option("-t, --topic").String("")
	c.payload = cli.Option("-p, --payload").String("")
	c.qos = cli.Option("-q, --qos").Uint16(0)
}

func (c *Pub) Run(ctx context.Context) error {
	conn, err := net.Dial("tcp", c.server.String())
	if err != nil {
		return err
	}

	// use middlewares and build your in/out queues with desired
	// features
	var (
		sender = tt.NewSender(conn)
		pool   = tt.NewIDPool(100)
		logger = tt.NewLogger(tt.LevelInfo)

		out  = tt.NewOutQueue(sender.Out, logger, pool)
		done = make(chan struct{}, 0)

		in = logger.In(pool.In(
			func(ctx context.Context, p mq.Packet) error {
				switch p.(type) {
				case *mq.ConnAck:
					{ // publish
						p := mq.NewPublish()
						p.SetQoS(uint8(c.qos))
						p.SetTopicName(c.topic)
						p.SetPayload([]byte(c.payload))
						return out(ctx, p)

					}

				case *mq.PubAck:
					// disconnect
					_ = out(ctx, mq.NewDisconnect())
					close(done)

				default:
					fmt.Println(p)
				}
				return nil
			},
		))
	)
	// start handling packet flow
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	failed := make(chan error, 0)
	go func() {
		if err := tt.NewReceiver(in, conn).Run(ctx); err != nil {
			if errors.Is(err, io.EOF) {
				failed <- fmt.Errorf("disconnected without ack")
			}
			close(failed)
		}
	}()

	p := mq.NewConnect()
	p.SetClientID("tt")
	_ = out(ctx, p)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	case err := <-failed:
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}
