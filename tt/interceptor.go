package tt

import (
	"context"

	"github.com/gregoryv/mq"
)

func Intercept[T mq.Packet]() *Interceptor[T] {
	c := make(chan T, 0)
	mid := func(next mq.Handler) mq.Handler {
		return func(ctx context.Context, p mq.Packet) error {
			switch p := p.(type) {
			case T:
				c <- p
			}
			return next(ctx, p)
		}
	}

	return &Interceptor[T]{
		InFunc: mid,
		c:      c,
	}
}

type Interceptor[T mq.Packet] struct {
	InFunc
	c chan T
}

func (i *Interceptor[T]) Done() <-chan T {
	return i.c
}
