package tt

import (
	"context"
	"sync"
	"time"

	"github.com/gregoryv/mq"
)

func NewSubWait(v int) *SubWait {
	return &SubWait{orig: v, count: v}
}

type SubWait struct {
	sync.Mutex
	orig  int
	count int
}

func (a *SubWait) In(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		switch p.(type) {
		case *mq.SubAck:
			a.Lock()
			a.count--
			a.Unlock()
		}
		return next(ctx, p)
	}
}

func (a *SubWait) Out(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		switch p.(type) {
		case *mq.Connect:
			a.reset()
		}
		return next(ctx, p)
	}
}

// AllSubscribed returns channel which blocks until expected number of
// mq.SubAck packets have been counted.
func (a *SubWait) AllSubscribed(ctx context.Context) <-chan struct{} {
	c := make(chan struct{})
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Millisecond):
				a.Lock()
				v := a.count
				a.Unlock()
				if v == 0 {
					c <- struct{}{}
					a.reset()
					return
				}
			}
		}
	}()
	return c
}

func (a *SubWait) reset() {
	a.Lock()
	a.count = a.orig
	a.Unlock()
}
