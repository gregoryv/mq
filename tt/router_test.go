package tt_test

import (
	"context"
	"fmt"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
)

func ExampleRouter() {
	routes := []*tt.Route{
		tt.NewRoute("gopher/pink", func(_ context.Context, p *mq.Publish) error {
			fmt.Println(p)
			return nil
		}),
		tt.NewRoute("gopher/blue", func(_ context.Context, p *mq.Publish) error {
			fmt.Println(p)
			return nil
		}),
	}
	r := tt.NewRouter()
	r.AddRoutes(routes...)

	ctx := context.Background()
	r.Route(ctx, Pub(0, "gopher/pink", "hi"))

	fmt.Print(r)
	//output:
	// PUBLISH ---- p0 20 bytes
	// 2 routes
}

func Pub(qos uint8, topic, payload string) *mq.Publish {
	p := mq.NewPublish()
	p.SetQoS(qos)
	p.SetTopicName(topic)
	p.SetPayload([]byte(payload))
	return &p
}
