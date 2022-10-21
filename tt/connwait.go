package tt

import (
	"context"

	"github.com/gregoryv/mq"
)

func NewConnWait() *ConnWait {
	return &ConnWait{
		c: make(chan struct{}, 0),
	}
}

type ConnWait struct {
	c chan struct{}
}

func (a *ConnWait) In(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		switch p.(type) {
		case *mq.ConnAck:
			a.c <- struct{}{}
		}
		return next(ctx, p)
	}
}

func (a *ConnWait) Done() <-chan struct{} {
	return a.c
}
