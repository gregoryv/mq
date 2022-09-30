package tt

import (
	"context"
	"fmt"

	"github.com/gregoryv/mq"
)

func NewSubscription(filter string, h mq.HandlerFunc) *Subscription {
	p := mq.NewSubscribe()
	p.AddFilter(filter, 0)

	return &Subscription{
		Subscribe: &p,
		Handler:   h,
	}
}

type Subscription struct {
	*mq.Subscribe
	mq.Handler
}

func missingHandler(_ context.Context, _ mq.Packet) error {
	return ErrMissingHandler
}

var ErrMissingHandler = fmt.Errorf("missing handler")
