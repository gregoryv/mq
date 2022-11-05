package tt

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"time"

	"github.com/gregoryv/mq"
)

// NewReceiver returns a receiver that reads packets from the reader
// and calls the handler.
func NewReceiver(h mq.Handler, r io.Reader) *Receiver {
	return &Receiver{
		wire:        r,
		handle:      h,
		readTimeout: 100 * time.Millisecond,
	}
}

type Receiver struct {
	wire        io.Reader
	handle      mq.Handler
	readTimeout time.Duration
}

// Run begins reading incoming packets and forwards them to the
// configured handler.
func (r *Receiver) Run(ctx context.Context) error {
loop:
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if w, ok := r.wire.(net.Conn); ok {
			w.SetReadDeadline(time.Now().Add(r.readTimeout))
		}
		p, err := mq.ReadPacket(r.wire)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue loop
			}
			return err
		}
		// ignore error here, it's up to the user to configure a queue
		// where the first middleware handles any errors, eg. Logger
		_ = r.handle(ctx, p)
	}
}

func Start(ctx context.Context, r Runner) <-chan error {
	c := make(chan error, 0)
	go func() {
		if err := r.Run(ctx); err != nil {
			if err != nil {
				c <- err
			}
			close(c)
		}
	}()
	return c
}

type Runner interface {
	Run(context.Context) error
}
