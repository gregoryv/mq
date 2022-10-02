package tt_test

import (
	"bytes"
	"context"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
)

var connection bytes.Buffer // replace with e.g. net.Conn

func Example_newClient() {
	c := tt.NewClient()
	c.SetIO(&connection)

	// start handling packet flow
	ctx, _ := context.WithCancel(context.Background())
	go c.Run(ctx)

	// connect
	p := mq.NewConnect()
	p.SetClientID("gopher")
	c.Connect(ctx, &p)
}
