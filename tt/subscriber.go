package tt

import (
	"context"

	"github.com/gregoryv/mq"
)

func NewSubscriber(send mq.Handler, routes ...*Route) *Subscriber {
	return &Subscriber{
		routes: routes,
		send:   send,
	}
}

type Subscriber struct {
	routes []*Route
	send   mq.Handler
}

func (s *Subscriber) SubscribeOnConnect(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		switch p.(type) {
		case *mq.ConnAck:
			// subscribe to each route separately, though you do not have to
			for _, r := range s.routes {
				_ = s.send(ctx, r.Subscribe())
			}
		}
		return next(ctx, p)
	}
}
