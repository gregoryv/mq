package tt

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/gregoryv/mq"
)

func NewNetClient(conn net.Conn) *Client {
	c := NewClient()
	c.SetReadWriter(conn)
	return c
}

func NewClient() *Client {
	maxConcurrentIds := uint16(100)
	c := &Client{
		debug:  log.New(log.Writer(), "", log.Flags()),
		ackman: NewAckman(NewIDPool(maxConcurrentIds)),
	}
	c.first = func(p mq.Packet) error {
		c.debug.Print(p)
		return nil
	}
	return c
}

type Client struct {
	m    sync.Mutex
	wire io.ReadWriter

	// todo
	first mq.Receiver

	ackman *Ackman
	debug  *log.Logger
}

func (c *Client) SetReadWriter(v io.ReadWriter) { c.wire = v }

// Run must be called before trying to send packets.
func (c *Client) Run(ctx context.Context) error {
	return c.handlePackets(ctx)
}

// Connect sends the packet and waits for acknowledgement. In the
// future this would be a good place to implement support for
// different auth methods.
func (c *Client) Connect(ctx context.Context, p *mq.Connect) error {
	c.setLogPrefix(p.ClientID())
	if err := c.send(p); err != nil {
		return fmt.Errorf("%w: %v", ErrConnect, err)
	}

	in, err := c.nextPacket() // todo move this to the chain of
	// receivers so we can intercept and use
	// shared logging
	if err != nil {
		return err
	}

	switch in := in.(type) {
	case *mq.ConnAck:
		c.setLogPrefix(in.AssignedClientID())
		if in.ReasonCode() != mq.Success {
			c.debug.Print("reason", in.ReasonString())
			return fmt.Errorf("%w: %s", ErrConnect, in.ReasonString())
		}

	default:
		c.debug.Print("unexpected", in)
		return fmt.Errorf("%w: unexpected %v", ErrConnect, in)
	}

	return nil
}

func (c *Client) Disconnect(ctx context.Context, p *mq.Disconnect) error {
	// todo handle session variations perhaps, async
	return c.debugErr(c.send(p))
}

func (c *Client) Pub(ctx context.Context, p *mq.Publish) error {
	if p.QoS() > 0 {
		id := c.ackman.Next(ctx)
		p.SetPacketID(id)
	}
	return c.debugErr(c.send(p))
}

func (c *Client) debugErr(err error) error {
	if err != nil {
		c.debug.Print(err)
	}
	return err
}

// Subscribe sends the subscribe packet to the connected broker.
// wip maybe introduce a subscription type
func (c *Client) Sub(ctx context.Context, p *mq.Subscribe) error {
	id := c.ackman.Next(ctx)
	p.SetPacketID(id)
	return c.debugErr(c.send(p))
}

// handlePackets is responsible for sending acks to incoming packets.
func (c *Client) handlePackets(ctx context.Context) error {
	for {
		in, err := c.nextPacket()
		if err != nil {
			c.debug.Print(err)
			c.debug.Print("no more packets will be handled")
			return err
		}

		// debug incoming control packet, todo move to the first receiver
		var buf bytes.Buffer
		in.WriteTo(&buf)
		msg := fmt.Sprint(in, " <- %s\n", hex.Dump(buf.Bytes()))

		select {
		case <-ctx.Done():
			return c.debugErr(ctx.Err())

		default:
			// reuse packet ids and handle acks
			switch in := in.(type) {
			case *mq.SubAck:
				// todo How will the ackman know what needs to be done
				// after ack ?  redesign this; as we need to possibly
				// notify caller, ie. if Subscribe is done in a sync
				// fashion
				c.ackman.Handle(ctx, in)

			case *mq.PubAck:
				c.ackman.Handle(ctx, in)

			default:
				msg = fmt.Sprintf(msg, "        (UNHANDLED!)")
			}
			msg = fmt.Sprintf(msg, "")
		}
		c.debug.Print(msg, "\n\n")
	}
}

// ----------------------------------------

func (c *Client) nextPacket() (mq.ControlPacket, error) {
	p, err := mq.ReadPacket(c.wire)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// send packet to the underlying connection.
func (c *Client) send(p mq.ControlPacket) error {
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
