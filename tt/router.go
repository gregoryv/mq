package tt

import (
	"context"
	"fmt"

	"github.com/gregoryv/mq"
)

// ----------------------------------------

func NewRouter() *Router {
	return &Router{}
}

type Router struct {
	routes []*Route
}

func (r *Router) String() string {
	return plural(len(r.routes), "route")
}

func (r *Router) Route(ctx context.Context, p *mq.Publish) error {
	// todo naive implementation looping over each route
	for _, route := range r.routes {
		if _, ok := route.Match(p.TopicName()); ok {
			for _, h := range route.handlers {
				_ = h(ctx, p) // todo how to handle errors
			}
		}
	}
	return ctx.Err()
}

func (r *Router) AddRoutes(routes ...*Route) error {
	r.routes = routes
	return nil
}

// ----------------------------------------

func plural(v int, word string) string {
	if v > 1 {
		word = word + "s"
	}
	return fmt.Sprintf("%v %s", v, word)
}

// Pub creates a new publish packet with the given values
func Pub(qos uint8, topicName, payload string) *mq.Publish {
	p := mq.NewPublish()
	p.SetQoS(qos)
	p.SetTopicName(topicName)
	p.SetPayload([]byte(payload))
	return &p
}
