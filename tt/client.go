package tt

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/gregoryv/mq"
)

// NewBasicClient returns a client with MaxDefaultConcurrentID and
// disabled logging
func NewBasicClient() *Client {
	fpool := NewPoolFeature(MaxDefaultConcurrentID)
	flog := NewLogFeature()

	c := NewClient()
	c.flog = flog // todo should not be needed but Settings requires it
	s := c.Settings()
	s.InStackSet([]mq.Middleware{
		flog.LogIncoming,
		flog.DumpPacket,
		fpool.ReusePacketID,
		flog.PrefixLoggers,
	})
	s.OutStackSet([]mq.Middleware{
		flog.PrefixLoggers,
		fpool.SetPacketID,
		flog.logOutgoing, // keep loggers last
		flog.DumpPacket,
	})
	return c
}

func NewClient() *Client {
	return &Client{
		// receiver should be replaced by the application layer
		receiver: unsetReceiver,
		out:      notRunning,
	}
}

type Client struct {
	flog *LogFeature

	running bool // set by func Run

	m    sync.Mutex
	wire io.ReadWriter

	// sequence of receivers for incoming packets
	instack  []mq.Middleware
	receiver mq.Handler // final

	outstack []mq.Middleware
	out      mq.Handler // first outgoing handler, set by func Run
}

func (c *Client) Start(ctx context.Context) {
	go c.Run(ctx)
	// wait for the run loop to be ready
	for {
		<-time.After(time.Millisecond)
		if c.running {
			<-time.After(5 * time.Millisecond)
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

	defer func() { c.running = false }()
	for {
		c.running = true
		p, err := c.nextPacket()
		if err != nil {
			// todo handle closed wire properly so clients may have
			// the feature of reconnect
			return err
		}
		if p != nil {
			incoming(ctx, p)
		}
	}
}

func stack(v []mq.Middleware, last mq.Handler) mq.Handler {
	if len(v) == 0 {
		return last
	}
	return v[0](stack(v[1:], last))
}

// Send the packet through the outgoing stack of handlers
func (c *Client) Send(ctx context.Context, p mq.Packet) error {
	return c.out(ctx, p)
}

// Settings returns this clients settings. If the client is running
// settings are read only.
func (c *Client) Settings() Settings {
	s := readSettings{c}
	if c.running {
		return &s
	}
	return &writeSettings{s}
}

func (c *Client) nextPacket() (mq.Packet, error) {
	p, err := mq.ReadPacket(c.wire)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// send packet to the underlying connection.
func (c *Client) send(_ context.Context, p mq.Packet) error {
	if c.wire == nil {
		return ErrNoConnection
	}
	c.m.Lock()
	_, err := p.WriteTo(c.wire)
	c.m.Unlock()
	return err
}

func unsetReceiver(_ context.Context, _ mq.Packet) error {
	return ErrUnsetReceiver
}

func notRunning(_ context.Context, _ mq.Packet) error {
	return ErrNotRunning
}
