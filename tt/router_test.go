package tt_test

import (
	"context"
	"fmt"

	"github.com/gregoryv/mq"
)

func ExampleRouter() {
	routes := []*Route{
		NewRoute("gopher/pink", func(ctx context.Context, p mq.Packet) error {
			switch p := p.(type) {
			case *mq.Publish:
				fmt.Println(p.TopicName(), string(p.Payload()))
			}
			return nil
		}),
	}
	r := NewRouter()
	r.AddRoutes(routes...)

	ctx := context.Background()
	r.Route(ctx, Pub(0, "gopher/pink", "hi"))

	fmt.Print(r)
	//output:
	// gopher/pink hi
	// 1 route
}

// ----------------------------------------

func Pub(qos uint8, topicName, payload string) *mq.Publish {
	p := mq.NewPublish()
	p.SetQoS(qos)
	p.SetTopicName(topicName)
	p.SetPayload([]byte(payload))
	return &p
}

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
		for _, r := range r.routes {
			if r.Match(p.TopicName()) {
				r.handler(ctx, p)
			}
		}
	}
	return ctx.Err()
}

func (r *Router) AddRoutes(routes ...*Route) error {
	r.routes = routes
	return fmt.Errorf("AddRoute: todo")
}

// ----------------------------------------

func NewRoute(filter string, h mq.Handler) *Route {
	return &Route{filter: filter, handler: h}
}

type Route struct {
	filter  string
	handler mq.Handler
}

func (r *Route) Match(name string) bool {

	if name == r.filter {
		return true
	}
	return false
}

func plural(v int, word string) string {
	if v > 1 {
		word = word + "s"
	}
	return fmt.Sprintf("%v %s", v, word)
}
