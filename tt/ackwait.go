package tt

import (
	"context"
	"sync"
	"time"

	"github.com/gregoryv/mq"
)

func NewAckWait(v int) *AckWait {
	return &AckWait{orig: v, count: v}
}

type AckWait struct {
	sync.Mutex
	orig  int
	count int
}

// Use resets on mq.Connect and counts mq.SubAck. Must be used in both
// in and out queues.
func (a *AckWait) Use(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		switch p.(type) {
		case *mq.Connect:
			a.reset()

		case *mq.SubAck:
			a.Lock()
			a.count--
			a.Unlock()
		}
		return next(ctx, p)
	}
}

// AllSubscribed returns channel which blocks until expected number of
// mq.SubAck packets have been counted.
func (a *AckWait) AllSubscribed(ctx context.Context) <-chan struct{} {
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

func (a *AckWait) reset() {
	a.Lock()
	a.count = a.orig
	a.Unlock()
}
