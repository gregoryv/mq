package tt_test

import (
	"context"
	"fmt"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
)

func ExampleRouter() {
	routes := []*tt.Route{
		tt.NewRoute("gopher/pink", func(ctx context.Context, p mq.Packet) error {
			switch p := p.(type) {
			case *mq.Publish:
				fmt.Println(p.TopicName(), string(p.Payload()))
			}
			return nil
		}),
	}
	r := tt.NewRouter()
	r.AddRoutes(routes...)

	ctx := context.Background()
	r.Route(ctx, tt.Pub(0, "gopher/pink", "hi"))

	fmt.Print(r)
	//output:
	// gopher/pink hi
	// 1 route
}

// ----------------------------------------
