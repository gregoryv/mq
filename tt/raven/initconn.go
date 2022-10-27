package raven

import (
	. "context"
	"io"

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

	_ = out // todo register outgoing connection once connected
	go tt.NewReceiver(conn, in).Run(ctx)

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
		if _, err := a.WriteTo(conn); err != nil {
			return "", err
		}
		return id, nil

	case <-ctx.Done():
		// stopped from the outside
		return "", ctx.Err()
	}

}
