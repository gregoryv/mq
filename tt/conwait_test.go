package tt

import (
	"context"
	"testing"
	"time"

	"github.com/gregoryv/mq"
)

func TestConnWait(t *testing.T) {
	conn, server := Dial()

	var (
		sender  = NewSender(conn).Out
		conwait = NewConnWait()
		logger  = NewLogger(LevelInfo)
	)

	out := NewOutQueue(sender, conwait, logger)

	in := NewInQueue(NoopHandler, conwait, logger)

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
	<-conwait.Done(ctx)
	if err := ctx.Err(); err != nil {
		t.Fatal(err)
	}
}
