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

func (r *Router) Route(ctx context.Context, p mq.Packet) error {
	switch p := p.(type) {
	case *mq.Publish:
		// todo naive implementation looping over each route
		for _, r := range r.routes {
			if _, ok := r.Match(p.TopicName()); ok {
				for _, h := range r.handlers {
					_ = h(ctx, p) // todo how to handle errors
				}
			}
		}
	}
	return ctx.Err()
}

func (r *Router) Add(filter string, handlers ...mq.PubHandler) error {
	r.routes = append(r.routes, NewRoute(filter, handlers...))
	return nil
}

func (r *Router) AddRoutes(routes ...*Route) error {
	r.routes = routes
	return fmt.Errorf("AddRoute: todo")
}

func (r *Router) Routes() []*Route {
	return r.routes
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
