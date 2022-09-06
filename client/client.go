package client

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/gregoryv/mqtt"
)

func NewClient(conn io.ReadWriter) *Client {

	c := &Client{
		ReadWriter: conn,
		debug:      log.New(log.Writer(), "", log.Flags()),
		pool:       NewIDPool(100),
	}
	return c
}

type Client struct {
	m sync.Mutex
	io.ReadWriter

	pool *IDPool

	debug *log.Logger
}

// Connect sends the packet and waits for acknoledgement. In the
// future this would be a good place to implement support for
// different auth methods.
func (c *Client) Connect(ctx context.Context, p *mqtt.Connect) error {
	c.setLogPrefix(p.ClientID())
	c.Send(p)

	in, err := c.nextPacket()
	if err != nil {
		return err
	}

	switch in := in.(type) {
	case *mqtt.ConnAck:
		c.setLogPrefix(in.AssignedClientID())
		if in.ReasonCode() != mqtt.Success {
			c.debug.Print("reason", in.ReasonString())
		}
	default:
		c.debug.Print("unexpected", in)
	}

	go c.handlePackets(ctx)
	return nil
}

func (c *Client) handlePackets(ctx context.Context) {
	for {
		in, err := c.nextPacket()
		if err != nil {
			c.debug.Print(err)
			c.debug.Print("no more packets will be handled")
			return
		}

		// debug incoming control packet
		var buf bytes.Buffer
		in.WriteTo(&buf)
		msg := fmt.Sprint(in, " <- %s\n", hex.Dump(buf.Bytes()))

		select {
		case <-ctx.Done():
			c.debug.Print(ctx.Err())
			return

		default:

			switch in := in.(type) {
			case *mqtt.SubAck:
				c.pool.Reuse(in.PacketID())

			case *mqtt.PubAck:
				c.pool.Reuse(in.PacketID())

			default:
				msg = fmt.Sprintf(msg, "        (UNHANDLED!)")
			}
			msg = fmt.Sprintf(msg, "")
		}
		c.debug.Print(msg, "\n\n")
	}
}

func (c *Client) Disconnect(p *mqtt.Disconnect) error {
	// todo handle session variations perhaps, async
	return c.Send(p)
}

func (c *Client) Publish(ctx context.Context, p *mqtt.Publish) {
	if err := c.publish(ctx, p); err != nil {
		c.debug.Print(err)
	}
}

func (c *Client) publish(ctx context.Context, p *mqtt.Publish) error {
	if p.QoS() > 0 {
		id := c.pool.Next(ctx)
		p.SetPacketID(id)
	}
	return c.Send(p)
}

func (c *Client) Subscribe(ctx context.Context, p *mqtt.Subscribe) error {
	// todo handle subscription, async
	id := c.pool.Next(ctx)
	p.SetPacketID(id)

	return c.Send(p)
}

// ----------------------------------------

func (c *Client) nextPacket() (mqtt.ControlPacket, error) {
	p, err := mqtt.ReadPacket(c)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Send packet to the underlying connection.
func (c *Client) Send(p mqtt.ControlPacket) error {
	// todo handle packet ids I guess
	c.m.Lock()
	_, err := p.WriteTo(c)
	c.m.Unlock()
	if err != nil {
		c.debug.Print("<- ", p, err)
		return err
	}
	var buf bytes.Buffer
	p.WriteTo(&buf)
	c.debug.Print("<- ", p, "\n", hex.Dump(buf.Bytes()), "\n")
	return nil
}

func (c *Client) setLogPrefix(cid string) {
	switch {
	case cid == "":
		return

	case len(cid) > 16:
		c.debug.SetPrefix(cid[len(cid)-8:] + "  ")

	default:
		c.debug.SetPrefix(cid + " ")
	}
}
