package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"time"

	"github.com/gregoryv/cmdline"
	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
)

type PubCmd struct {
	server *url.URL

	topic    string
	payload  string
	qos      uint8
	timeout  time.Duration
	clientID string
}

func (c *PubCmd) ExtraOptions(cli *cmdline.Parser) {
	c.server = cli.Option("-s, --server").Url("localhost:1883")
	c.topic = cli.Option("-t, --topic").String("gopher/pink")
	c.payload = cli.Option("-p, --payload").String("hug")
	c.qos = cli.Option("-q, --qos").Uint8(0)
	c.timeout = cli.Option("--timeout").Duration("1s")
	c.clientID = cli.Option("-cid, --client-id").String("ttpub")
}

func (c *PubCmd) Run(ctx context.Context) error {
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

		out     = tt.NewOutQueue(sender.Out, logger, pool)
		done    = make(chan struct{}, 0) // closed by handler on success
		handler mq.Handler
		msg     = mq.Pub(c.qos, c.topic, c.payload)
	)
	logger.SetLogPrefix(c.clientID)

	// QoS dictates the logic of packet flows
	switch c.qos {
	case 0:
		handler = func(ctx context.Context, p mq.Packet) error {
			switch p.(type) {
			case *mq.ConnAck:
				if err := out(ctx, msg); err != nil {
					return err
				}
				close(done)
			default:
				fmt.Println("unexpected:", p)
			}
			return nil
		}
	case 1:
		handler = func(ctx context.Context, p mq.Packet) error {
			switch p.(type) {
			case *mq.ConnAck:
				return out(ctx, msg)
			case *mq.PubAck:
				close(done)
			default:
				fmt.Println("unexpected:", p)
			}
			return nil
		}

	case 2:
		handler = func(ctx context.Context, p mq.Packet) error {
			switch p := p.(type) {
			case *mq.ConnAck:
				return out(ctx, msg)
			case *mq.PubAck:
				switch p.AckType() {
				case mq.PUBREC:
					rel := mq.NewPubRel()
					rel.SetPacketID(msg.PacketID())
					return out(ctx, rel)
				case mq.PUBCOMP:
					close(done)
				}
			default:
				fmt.Println("unexpected:", p)
			}
			return nil
		}

	default:
		return fmt.Errorf("cannot handle QoS %v", c.qos)
	}

	var (
		in       = tt.NewInQueue(handler, pool, logger)
		receiver = tt.NewReceiver(in, conn)
	)
	// start handling packet flow
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	running := tt.Start(ctx, receiver)

	// kick off with a connect

	p := mq.NewConnect()
	p.SetClientID(c.clientID)
	_ = out(ctx, p)

	select {
	case err := <-running:
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("FAIL")
		}

	case <-ctx.Done():
		return ctx.Err()

	case <-done:
		defer fmt.Println("ok")
	}
	_ = out(ctx, mq.NewDisconnect())
	return nil
}
