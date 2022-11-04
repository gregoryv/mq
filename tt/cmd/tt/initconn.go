package main

import (
	"context"
	. "context"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
)

// todo maybe a connections handler of sorts that keeps track of
// unique connections

// InitConn returns the client id after a successful connect and
// ack.
func InitConn(ctx Context, conn io.ReadWriter) (string, error) {
	var (
		sender    = tt.NewSender(conn)
		onConnect = make(chan *mq.Connect, 0)
		connwait  = tt.Intercept(onConnect)
		logger    = NewLogger(tt.LevelInfo)

		in  = tt.NewInQueue(tt.NoopHandler, connwait, logger)
		out = tt.NewOutQueue(sender.Out, logger)
	)
	defer close(onConnect)

	connectTimeout := time.Second // duration until the first Connect packet comes in
	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	go tt.NewReceiver(in, conn).Run(ctx)

	select {
	case p := <-onConnect:
		// connect came in...
		a := mq.NewConnAck()
		id := p.ClientID()
		if id == "" {
			id = uuid.NewString()
		}
		// todo make sure it's uniq
		a.SetAssignedClientID(id)
		cancel()
		return id, out(ctx, a)

	case <-ctx.Done():
		// stopped from the outside
		return "", ctx.Err()
	}

}
