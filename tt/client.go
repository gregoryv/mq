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

// todo maybe this is more like a port that can be used both on client
// and server side it's first when you set the application receiver
// that they differ, right?

func NewNetClient(conn net.Conn) *Client {
	c := NewClient()
	c.SetReadWriter(conn)
	return c
}

func NewClient() *Client {
	maxConcurrentIds := uint16(100)
	c := &Client{
		debug:    log.New(log.Writer(), "", log.Flags()),
		ackman:   NewAckman(NewIDPool(maxConcurrentIds)),
		Incoming: make(chan mq.ControlPacket, 0),
	}
	c.first = c.debugPacket(
		c.interceptPacket(
			c.handleAckPacket,
		),
	)
	return c
}

type Client struct {
	m    sync.Mutex
	wire io.ReadWriter

	// todo use it in handlePackets
	first    mq.Receiver
	Incoming chan mq.Packet // allows for intercepting packets

	ackman *Ackman
	debug  *log.Logger
}

func (c *Client) debugPacket(next mq.Receiver) mq.Receiver {
	return func(p mq.Packet) error {
		var buf bytes.Buffer
		p.WriteTo(&buf)
		msg := fmt.Sprint(p, " <- %s\n", hex.Dump(buf.Bytes()))
		c.debug.Printf(msg, "")
		c.debug.Print("\n\n")

		return next(p)
	}
}

func (c *Client) interceptPacket(next mq.Receiver) mq.Receiver {
	return func(p mq.Packet) error {
		select {
		case c.Incoming <- p:
		default:
		}
		return next(p)
	}
}

func (c *Client) handleAckPacket(p mq.Packet) error {
	ctx := context.Background()
	// reuse packet ids and handle acks
	switch p := p.(type) {
	case *mq.SubAck:
		c.ackman.Handle(ctx, p) // todo move to first or subsequent, why?

	case *mq.PubAck:
		c.ackman.Handle(ctx, p)

	case *mq.Publish:
		c.ackman.Handle(ctx, p)

	case *mq.ConnAck:
		c.setLogPrefix(p.AssignedClientID())
		if p.ReasonCode() != mq.Success {
			c.debug.Print("reason", p.ReasonString())
		}
	}
	return nil
}

func (c *Client) SetReadWriter(v io.ReadWriter) { c.wire = v }

// Connect sends the packet and waits for acknowledgement. In the
// future this would be a good place to implement support for
// different auth methods.
func (c *Client) Connect(ctx context.Context, p *mq.Connect) error {
	c.setLogPrefix(p.ClientID())
	return c.debugErr(c.send(p))
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

// Run must be called before trying to send packets.
func (c *Client) Run(ctx context.Context) error {
	return c.handlePackets(ctx)
}

// handlePackets is responsible for sending acks to incoming packets.
func (c *Client) handlePackets(ctx context.Context) error {
	for {
		p, err := c.nextPacket()
		if err != nil {
			c.debug.Print(err)
			c.debug.Print("no more packets will be handled")
			return err
		}
		if p != nil {
			c.first(p)
		}
	}
}

// ----------------------------------------

func (c *Client) nextPacket() (mq.Packet, error) {
	p, err := mq.ReadPacket(c.wire)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// send packet to the underlying connection.
func (c *Client) send(p mq.Packet) error {
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
