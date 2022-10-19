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
)

func NewQueue() *Queue {
	return &Queue{
		// receiver should be replaced by the application layer
		receiver:    unsetReceiver,
		out:         notRunning,
		readTimeout: 100 * time.Millisecond,
	}
}

type Queue struct {
	running bool // set by func Run

	m           sync.Mutex
	wire        io.ReadWriter
	readTimeout time.Duration

	// sequence of receivers for incoming packets
	incoming mq.Handler // set by func Run
	instack  []mq.Middleware
	receiver mq.Handler // final

	outstack []mq.Middleware
	out      mq.Handler // first outgoing handler, set by func Run
}

func (q *Queue) Start(ctx context.Context) {
	go q.Run(ctx)
	// wait for the run loop to be ready
	for {
		<-time.After(time.Millisecond)
		if q.running {
			<-time.After(5 * time.Millisecond)
			return
		}
	}
}

// Send the packet through the outgoing idpool of handlers
func (q *Queue) Send(ctx context.Context, p mq.Packet) error {
	return q.out(ctx, p)
}

// Send the packet through the outgoing idpool of handlers
func (q *Queue) Recv(ctx context.Context, p mq.Packet) error {
	return q.incoming(ctx, p)
}

// Run begins handling incoming packets and must be called before
// trying to send packets. Run blocks until context is interrupted,
// the wire has closed or there a malformed packet is encountered.
func (q *Queue) Run(ctx context.Context) error {
	q.incoming = chain(q.instack, q.receiver)
	q.out = chain(q.outstack, q.send)

	defer func() { q.running = false }()
	for {
		q.running = true
		if w, ok := q.wire.(net.Conn); ok {
			w.SetReadDeadline(time.Now().Add(q.readTimeout))
		}
		p, err := q.nextPacket()
		if err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
			// todo handle closed wire properly so clients may have
			// the feature of reconnect
			return err
		}
		if p != nil {
			// ignore error here, it's up to the user to configure a
			// stack where the first middleware handles any errors.
			_ = q.Recv(ctx, p)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
}
func (q *Queue) InStackSet(v []mq.Middleware) error {
	if q.running {
		return ErrReadOnly
	}
	q.instack = v
	return nil
}

func (q *Queue) OutStackSet(v []mq.Middleware) error {
	if q.running {
		return ErrReadOnly
	}
	q.outstack = v
	return nil
}

func (q *Queue) ReceiverSet(v mq.Handler) error {
	if q.running {
		return ErrReadOnly
	}
	q.receiver = v
	return nil
}

func (q *Queue) IOSet(v io.ReadWriter) error {
	if q.running {
		return ErrReadOnly
	}
	q.wire = v
	return nil
}

func chain(v []mq.Middleware, last mq.Handler) mq.Handler {
	if len(v) == 0 {
		return last
	}
	return v[0](chain(v[1:], last))
}

// nextPacket reads from the configured IO with a timeout
func (q *Queue) nextPacket() (mq.Packet, error) {
	p, err := mq.ReadPacket(q.wire)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// send writes a packet to the underlying connection.
func (q *Queue) send(_ context.Context, p mq.Packet) error {
	if q.wire == nil {
		return ErrNoConnection
	}
	q.m.Lock()
	_, err := p.WriteTo(q.wire)
	q.m.Unlock()
	return err
}

func unsetReceiver(_ context.Context, _ mq.Packet) error {
	return ErrUnsetReceiver
}

func notRunning(_ context.Context, _ mq.Packet) error {
	return ErrNotRunning
}
