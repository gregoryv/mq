package tt

import (
	"context"
	"testing"
	"time"

	"github.com/gregoryv/mq"
)

func TestSubWait(t *testing.T) {
	conn, server := Dial()

	routes := []*Route{
		NewRoute("#"),
		NewRoute("a/b"),
	}

	var (
		sender  = NewSender(conn).Out
		subwait = NewSubWait(len(routes))

		out = NewOutQueue(sender, subwait)
		in  = NewInQueue(NoopHandler, subwait)
	)

	// start handling packet flow
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	go NewReceiver(conn, in).Run(ctx)

	for _, r := range routes {
		p := mq.NewSubscribe()
		p.AddFilter(r.Filter(), mq.OptNL)
		_ = out(ctx, &p)
		server.Ack(&p)
	}

	<-subwait.Done(ctx)
	if err := ctx.Err(); err != nil {
		t.Error(err)
	}

	p := mq.NewSubscribe()
	_ = out(ctx, &p)
	// without server ack
	select {
	case <-subwait.Done(ctx):
		t.Error("AllSubscribed should timeout")
	case <-time.After(time.Millisecond):
	}
}
