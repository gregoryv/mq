package tt

import (
	"context"
	"strings"
	"testing"

	"github.com/gregoryv/mq"
)

func TestRouter(t *testing.T) {
	routes := []*Route{
		NewRoute("gopher/pink", NoopPub),
		NewRoute("gopher/blue", NoopPub),
		NewRoute("#", NoopPub),
	}
	r := NewRouter(routes...)

	ctx := context.Background()
	if err := r.In(ctx, mq.Pub(0, "gopher/pink", "hi")); err != nil {
		t.Error(err)
	}

	if v := r.String(); !strings.Contains(v, "3 routes") {
		t.Error(v)
	}
}
