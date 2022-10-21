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

func NewReceiver(r io.Reader, first mq.Handler) *Receiver {
	return &Receiver{
		wire:        r,
		first:       first,
		readTimeout: 100 * time.Millisecond,
	}
}

type Receiver struct {
	wire        io.Reader
	first       mq.Handler
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
		_ = r.first(ctx, p)
	}
}
