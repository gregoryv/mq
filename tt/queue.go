/*
Package tt provides components for writing mqtt-v5 clients and servers.
*/
package tt

import (
	"context"

	"github.com/gregoryv/mq"
)

func NewInQueue(last mq.Handler, v ...InFlow) mq.Handler {
	if len(v) == 0 {
		return last
	}
	l := len(v) - 1
	return v[l].In(NewInQueue(last, v[:l]...))
}

func NewOutQueue(last mq.Handler, v ...OutFlow) mq.Handler {
	if len(v) == 0 {
		return last
	}
	l := len(v) - 1
	return v[l].Out(NewOutQueue(last, v[:l]...))
}

type InOutFlow interface {
	InFlow
	OutFlow
}

type InFlow interface {
	In(next mq.Handler) mq.Handler
}

type OutFlow interface {
	Out(next mq.Handler) mq.Handler
}

func NoopHandler(_ context.Context, _ mq.Packet) error { return nil }
func NoopPub(_ context.Context, _ *mq.Publish) error   { return nil }

type InFunc func(mq.Handler) mq.Handler

func (f InFunc) In(next mq.Handler) mq.Handler {
	return f(next)
}

type OutFunc func(mq.Handler) mq.Handler

func (f OutFunc) Out(next mq.Handler) mq.Handler {
	return f(next)
}
