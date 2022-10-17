// Package intercept provides a package interceptor
package intercept

import (
	"context"
	"time"

	"github.com/gregoryv/mq"
)

// New returns a new interceptor with max buffer of intercepte packages
func New(max int) *Interceptor {
	return &Interceptor{
		C: make(chan mq.Packet, max),
	}
}

type Interceptor struct {
	C chan mq.Packet
}

// Intercept adds the incoming packet to the channel, if channel is
// full it continues with the next handler after 10ms.
func (r *Interceptor) Intercept(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		select {
		case r.C <- p: // if anyone is interested
		case <-time.After(10 * time.Millisecond):
		}
		return next(ctx, p)
	}
}
