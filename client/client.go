package client

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/gregoryv/mqtt"
)

func NewNetClient(conn net.Conn) *Client {
	c := NewClient()
	c.SetReadWriter(conn)
	return c
}

func NewClient() *Client {
	c := &Client{
		debug:  log.New(log.Writer(), "", log.Flags()),
		ackman: NewAckman(NewIDPool(100)),
	}
	return c
}

// todo what is the purpose of the client?
type Client struct {
	m    sync.Mutex
	wire io.ReadWriter

	ackman *Ackman
	debug  *log.Logger
}

func (c *Client) SetReadWriter(v io.ReadWriter) { c.wire = v }

// Connect sends the packet and waits for acknowledgement. In the
// future this would be a good place to implement support for
// different auth methods.
func (c *Client) Connect(ctx context.Context, p *mqtt.Connect) error {
	c.setLogPrefix(p.ClientID())
	if err := c.send(p); err != nil {
		return fmt.Errorf("%w: %v", ErrConnect, err)
	}

	in, err := c.nextPacket()
	if err != nil {
		return err
	}

	switch in := in.(type) {
	case *mqtt.ConnAck:
		c.setLogPrefix(in.AssignedClientID())
		if in.ReasonCode() != mqtt.Success {
			c.debug.Print("reason", in.ReasonString())
			return fmt.Errorf("%w: %s", ErrConnect, in.ReasonString())
		}

	default:
		c.debug.Print("unexpected", in)
		return fmt.Errorf("%w: unexpected %v", ErrConnect, in)
	}

	go c.handlePackets(ctx)
	return nil
}

// handlePackets is responsible for sending acks to incoming packets.
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
			// reuse packet ids and handle acks
			switch in := in.(type) {
			case *mqtt.SubAck:
				// todo How will the ackman know what needs to be done
				// after ack ?  redesign this; as we need to possibly
				// notify caller, ie. if Subscribe is done in a sync
				// fashion
				c.ackman.Handle(ctx, in)

			case *mqtt.PubAck:
				c.ackman.Handle(ctx, in)

			default:
				msg = fmt.Sprintf(msg, "        (UNHANDLED!)")
			}
			msg = fmt.Sprintf(msg, "")
		}
		c.debug.Print(msg, "\n\n")
	}
}

func (c *Client) Disconnect(p *mqtt.Disconnect) {
	// todo handle session variations perhaps, async
	if err := c.send(p); err != nil {
		c.debug.Print(err)
	}
}

func (c *Client) Publish(ctx context.Context, p *mqtt.Publish) {
	if err := c.publish(ctx, p); err != nil {
		c.debug.Print(err)
	}
}

func (c *Client) publish(ctx context.Context, p *mqtt.Publish) error {
	if p.QoS() > 0 {
		id := c.ackman.Next(ctx)
		p.SetPacketID(id)
	}
	return c.send(p)
}

func (c *Client) Subscribe(ctx context.Context, p *mqtt.Subscribe) error {
	// todo handle subscription, async
	id := c.ackman.Next(ctx)
	p.SetPacketID(id)

	return c.send(p)
}

// ----------------------------------------

func (c *Client) nextPacket() (mqtt.ControlPacket, error) {
	p, err := mqtt.ReadPacket(c.wire)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// send packet to the underlying connection.
func (c *Client) send(p mqtt.ControlPacket) error {
	if c.wire == nil {
		return ErrNoConnection
	}
	// todo handle packet ids I guess
	c.m.Lock()
	_, err := p.WriteTo(c.wire)
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
		c.debug.SetPrefix("          ")
		return

	case len(cid) > 16:
		c.debug.SetPrefix(cid[len(cid)-8:] + "  ")

	default:
		c.debug.SetPrefix(cid + " ")
	}
}

var (
	ErrNoConnection = fmt.Errorf("no connection")
	ErrConnect      = fmt.Errorf("connect error")
)
