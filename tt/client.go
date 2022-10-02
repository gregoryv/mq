package tt

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/gregoryv/mq"
)

func NewClient() *Client {
	maxConcurrentIds := uint16(100)
	c := &Client{
		debug: log.New(log.Writer(), "", log.Flags()),
		pool:  newPool(maxConcurrentIds),
	}
	// sequence of receivers for incoming packets
	c.first = c.debugPacket(c.handleAckPacket(
		// final step forwards to the configured receiver
		func(p mq.Packet) error {
			return c.receiver(p)
		},
	))
	c.receiver = func(_ mq.Packet) error { return ErrUnsetReceiver }
	return c
}

type Client struct {
	m    sync.Mutex
	wire io.ReadWriter

	first    mq.Receiver
	receiver mq.Receiver // the application layer

	pool  *pool
	debug *log.Logger
}

// SetIO sets the read writer used for serializing packets from and to.
// Should be set before calling Run
func (c *Client) SetIO(v io.ReadWriter) { c.wire = v }

func (c *Client) SetReceiver(v mq.Receiver) { c.receiver = v }
func (c *Client) Receiver() mq.Receiver     { return c.receiver }

// Run begins handling incoming packets and must be called before
// trying to send packets. Run blocks until context is interrupted,
// the wire has closed or there a malformed packet is encountered.
func (c *Client) Run(ctx context.Context) error {
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

// Connect sends the packet. In the future this would be a good place
// to implement support for different auth methods.
func (c *Client) Connect(ctx context.Context, p *mq.Connect) error {
	c.setLogPrefix(p.ClientID())
	return c.debugErr(c.send(p))
}

func (c *Client) Disconnect(ctx context.Context, p *mq.Disconnect) error {
	// todo handle session variations perhaps, async
	return c.debugErr(c.send(p))
}

// Pub sends the packet and is safe for concurrent use by multiple
// goroutines.
func (c *Client) Pub(ctx context.Context, p *mq.Publish) error {
	if p.QoS() > 0 {
		id := c.pool.Next(ctx)
		p.SetPacketID(id)
	}
	return c.debugErr(c.send(p))
}

// Sub sends the packet and is safe for concurrent use by multiple
// goroutines.
func (c *Client) Sub(ctx context.Context, p *mq.Subscribe) error {
	id := c.pool.Next(ctx)
	p.SetPacketID(id)
	return c.debugErr(c.send(p))
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

func (c *Client) handleAckPacket(next mq.Receiver) mq.Receiver {
	return func(p mq.Packet) error {
		// reuse packet ids and handle acks
		switch p := p.(type) {
		case *mq.Publish:
			c.pool.Reuse(p.PacketID())

		case *mq.PubAck:
			c.pool.Reuse(p.PacketID())

		case *mq.SubAck:
			c.pool.Reuse(p.PacketID())

		case *mq.ConnAck:
			c.setLogPrefix(p.AssignedClientID())
			if p.ReasonCode() != mq.Success {
				c.debug.Print("reason", p.ReasonString())
			}
		}
		return next(p)
	}
}

func (c *Client) debugErr(err error) error {
	if err != nil {
		c.debug.Print(err)
	}
	return err
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
	ErrNoConnection  = fmt.Errorf("no connection")
	ErrConnect       = fmt.Errorf("connect error")
	ErrUnsetReceiver = fmt.Errorf("unset receiver")
)
