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
	}
	r := tt.NewRouter()
	r.AddRoutes(routes...)

	ctx := context.Background()
	r.Route(ctx, tt.Pub(0, "gopher/pink", "hi"))

	fmt.Print(r)
	//output:
	// PUBLISH ---- p0 20 bytes
	// 1 route
}
