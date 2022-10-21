/*
Package tt provides components for writing mqtt-v5 clients and servers.
*/
package tt

import (
	"context"

	"github.com/gregoryv/mq"
)

func NoopHandler(_ context.Context, _ mq.Packet) error { return nil }
func NoopPub(_ context.Context, _ *mq.Publish) error   { return nil }

type InFlow interface {
	In(next mq.Handler) mq.Handler
}
type OutFlow interface {
	Out(next mq.Handler) mq.Handler
}
