package tt

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"github.com/gregoryv/mq"
)

// NewClient returns a client with MaxDefaultConcurrentID and disabled logging
func NewClient() *Client {
	c := &Client{
		pool:  newPool(MaxDefaultConcurrentID),
		info:  log.New(log.Writer(), "", log.Flags()),
		debug: log.New(log.Writer(), "", log.Flags()),

		// this receiver should be replaced by the application layer
		receiver: func(_ mq.Packet) error { return ErrUnsetReceiver },
		out:      func(p mq.Packet) error { panic(p.String()) },
	}
	c.instack = []mq.Middleware{
		c.logIncoming,
		c.handleAckPacket,
	}
	c.outstack = []mq.Middleware{
		c.logOutgoing,
	}
	c.LogLevelSet(LogLevelNone)
	return c
}

type Client struct {
	pool  *pool // of packet IDs
	info  *log.Logger
	debug *log.Logger

	m    sync.Mutex
	wire io.ReadWriter

	// sequence of receivers for incoming packets
	instack  []mq.Middleware
	receiver mq.Handler // final

	outstack []mq.Middleware
	out      mq.Handler // first outgoing handler
}

// IOSet sets the read writer used for serializing packets from and to.
// Should be set before calling Run
func (c *Client) IOSet(v io.ReadWriter) { c.wire = v }

// ReceiverSet configures receiver for any incoming mq.Publish
// packets. The client handles PacketID reuse.
func (c *Client) ReceiverSet(v mq.Handler) { c.receiver = v }

// Receiver returns receiver setting.
func (c *Client) Receiver() mq.Handler { return c.receiver }

func (c *Client) LogLevelSet(v LogLevel) {
	switch v {
	case LogLevelDebug:
		c.info.SetOutput(log.Writer())
		c.debug.SetOutput(log.Writer())

	case LogLevelInfo:
		c.info.SetOutput(log.Writer())
		c.debug.SetOutput(ioutil.Discard)

	case LogLevelNone:
		c.info.SetOutput(ioutil.Discard)
		c.debug.SetOutput(ioutil.Discard)
	}
}

func (c *Client) Start(ctx context.Context) {
	go c.Run(ctx)
	// wait for the run loo to be ready
	for {
		<-time.After(time.Millisecond)
		if c.out != nil {
			time.After(5 * time.Millisecond)
			return
		}
	}
}

// Run begins handling incoming packets and must be called before
// trying to send packets. Run blocks until context is interrupted,
// the wire has closed or there a malformed packet is encountered.
func (c *Client) Run(ctx context.Context) error {
	incoming := stack(c.instack, c.receiver)
	c.out = stack(c.outstack, c.send)

	for {
		p, err := c.nextPacket()
		if err != nil {
			c.debug.Print(err)
			c.debug.Print("no more packets will be handled")
			return err
		}
		if p != nil {
			incoming(p)
		}
	}
}

func stack(v []mq.Middleware, last mq.Handler) mq.Handler {
	if len(v) == 0 {
		return last
	}
	return v[0](stack(v[1:], last))
}

// Connect sends the packet. In the future this would be a good place
// to implement support for different auth methods.
func (c *Client) Connect(ctx context.Context, p *mq.Connect) error {
	cid := p.ClientIDShort()
	c.setLogPrefix(cid)
	return c.out(p)
}

func (c *Client) Disconnect(ctx context.Context, p *mq.Disconnect) error {
	// todo handle session variations perhaps, async
	return c.out(p)
}

// Pub sends the packet and is safe for concurrent use by multiple
// goroutines. The packet ID is set if QoS > 0.
func (c *Client) Pub(ctx context.Context, p *mq.Publish) error {
	if p.QoS() > 0 {
		id := c.pool.Next(ctx)
		p.SetPacketID(id) // MQTT-2.2.1-3
	}
	return c.out(p)
}

// Sub sends the packet and is safe for concurrent use by multiple
// goroutines. Configure receiver using SetReceiver.
func (c *Client) Sub(ctx context.Context, p *mq.Subscribe) error {
	id := c.pool.Next(ctx)
	p.SetPacketID(id) // MQTT-2.2.1-3
	return c.out(p)
}

// Unsub sends the packet and is safe for concurrent use by multiple
// goroutines. Configure receiver using SetReceiver.
func (c *Client) Unsub(ctx context.Context, p *mq.Unsubscribe) error {
	id := c.pool.Next(ctx)
	p.SetPacketID(id) // MQTT-2.2.1-3
	return c.out(p)
}

func (c *Client) handleAckPacket(next mq.Handler) mq.Handler {
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
	return err
}

func (c *Client) logIncoming(next mq.Handler) mq.Handler {
	return func(p mq.Packet) error {
		c.debug.Print(p, " <- wire")
		var buf bytes.Buffer
		p.WriteTo(&buf)
		c.debug.Print("\n", hex.Dump(buf.Bytes()), "\n")
		return next(p)
	}
}

func (c *Client) logOutgoing(next mq.Handler) mq.Handler {
	return func(p mq.Packet) error {
		if err := next(p); err != nil {
			c.info.Print("wire <- ", p, err)
			return err
		}

		c.info.Print("wire <- ", p)
		var buf bytes.Buffer
		p.WriteTo(&buf)
		c.debug.Print("\n", hex.Dump(buf.Bytes()), "\n")
		return nil
	}
}

func (c *Client) setLogPrefix(cid string) {
	c.debug.SetPrefix(fmt.Sprintf("%s ", cid))
}

var (
	ErrNoConnection  = fmt.Errorf("no connection")
	ErrUnsetReceiver = fmt.Errorf("unset receiver")
)
