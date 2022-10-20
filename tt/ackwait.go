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

func (a *AckWait) CountSubAck(next mq.Handler) mq.Handler {
	return a.ResetOnConnect(next)
}

func (a *AckWait) ResetOnConnect(next mq.Handler) mq.Handler {
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
