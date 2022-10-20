package tt

import (
	"context"
	"time"

	"github.com/gregoryv/mq"
)

func NewConnWait() *ConnWait {
	return &ConnWait{}
}

type ConnWait struct {
	connected bool
}

func (a *ConnWait) In(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		switch p.(type) {
		case *mq.ConnAck:
			a.connected = true
		}
		return next(ctx, p)
	}
}

// Use resets on mq.Connect and counts mq.ConnAck. Must be used in both
// in and out queues.
func (a *ConnWait) Out(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		switch p.(type) {
		case *mq.Connect:
			a.connected = false
		}
		return next(ctx, p)
	}
}

// AllConnscribed returns channel which blocks until expected number of
// mq.ConnAck packets have been counted.
func (a *ConnWait) Done(ctx context.Context) <-chan struct{} {
	c := make(chan struct{})
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if a.connected {
					select {
					case c <- struct{}{}:
					case <-time.After(2 * time.Millisecond):
					}
					return
				}
			}
		}
	}()
	return c
}
