package pakio

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
	for {
		if w, ok := r.wire.(net.Conn); ok {
			w.SetReadDeadline(time.Now().Add(r.readTimeout))
		}
		p, err := mq.ReadPacket(r.wire)
		if err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
			return err
		}
		if p != nil {
			// ignore error here, it's up to the user to configure a
			// stack where the first middleware handles any errors.
			_ = r.first(ctx, p)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
}
