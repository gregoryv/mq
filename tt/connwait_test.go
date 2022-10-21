package tt

import (
	"context"
	"testing"
	"time"

	"github.com/gregoryv/mq"
)

func TestConnWait(t *testing.T) {
	var (
		conn, server = Dial()

		sender  = NewSender(conn).Out
		conwait = NewConnWait()

		out = NewOutQueue(sender)
		in  = NewInQueue(NoopHandler, conwait)
	)

	// start handling packet flow
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	go NewReceiver(conn, in).Run(ctx)

	{ // connect
		p := mq.NewConnect()
		p.SetClientID("connwait-test")
		_ = out(ctx, &p)
		server.Ack(&p)
	}

	<-conwait.Done()
	if err := ctx.Err(); err != nil {
		t.Fatal(err)
	}

	// times out
	{ // connect
		p := mq.NewConnect()
		p.SetClientID("connwait-timeout")
		_ = out(ctx, &p)
	}

	select {
	case <-conwait.Done():
		t.Error("Done should timeout")
	case <-time.After(time.Millisecond):
	}
}
