package tt

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt/flog"
	"github.com/gregoryv/mq/tt/idpool"
)

// NewBasicClient returns a client with MaxDefaultConcurrentID and
// disabled logging
func NewBasicClient() *Client {
	fpool := idpool.New(10)
	fl := flog.New()

	c := NewClient()
	s := c.Settings()
	s.InStackSet([]mq.Middleware{
		fl.LogIncoming,
		fl.DumpPacket,
		fpool.ReusePacketID,
		fl.PrefixLoggers,
	})
	s.OutStackSet([]mq.Middleware{
		fl.PrefixLoggers,
		fpool.SetPacketID,
		fl.LogOutgoing,
		fl.DumpPacket,
	})
	return c
}

func NewClient() *Client {
	return &Client{
		// receiver should be replaced by the application layer
		receiver:    unsetReceiver,
		out:         notRunning,
		readTimeout: 100 * time.Millisecond,
	}
}

type Client struct {
	running bool // set by func Run

	m           sync.Mutex
	wire        io.ReadWriter
	readTimeout time.Duration

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
	incoming := chain(c.instack, c.receiver)
	c.out = chain(c.outstack, c.send)

	defer func() { c.running = false }()
	for {
		c.running = true
		if w, ok := c.wire.(net.Conn); ok {
			w.SetReadDeadline(time.Now().Add(c.readTimeout))
		}
		p, err := c.nextPacket()
		if err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
			// todo handle closed wire properly so clients may have
			// the feature of reconnect
			return err
		}
		if p != nil {
			// ignore error here, it's up to the user to configure a
			// stack where the first middleware handles any errors.
			_ = incoming(ctx, p)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
}

func chain(v []mq.Middleware, last mq.Handler) mq.Handler {
	if len(v) == 0 {
		return last
	}
	return v[0](chain(v[1:], last))
}

// Send the packet through the outgoing idpool of handlers
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

// nextPacket reads from the configured IO with a timeout
func (c *Client) nextPacket() (mq.Packet, error) {
	p, err := mq.ReadPacket(c.wire)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// send writes a packet to the underlying connection.
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
