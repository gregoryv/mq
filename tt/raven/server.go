package raven

import (
	. "context"
	"errors"
	"io"
	"net"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt"
)

func NewServer() *Server {
	return &Server{
		bind:           ":", // random
		acceptTimeout:  time.Millisecond,
		connectTimeout: 20 * time.Millisecond,
		clients:        make(map[string]io.ReadWriter),
	}
}

type Server struct {
	bind string

	acceptTimeout time.Duration

	// client has to send the initial connect packet
	connectTimeout time.Duration

	clients map[string]io.ReadWriter
}

// Run listens for tcp connections. Blocks until context is cancelled
// or accepting a connection fails. Accepting new connection can only
// be interrupted if listener has SetDeadline method.
func (s *Server) Run(l net.Listener, ctx Context) error {
loop:
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Wait for a connection.
		if l, ok := l.(interface{ SetDeadline(time.Time) error }); ok {
			l.SetDeadline(time.Now().Add(s.acceptTimeout))
		}
		conn, err := l.Accept()
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue loop
			}
			// todo check what causes Accept to fail other than
			// timeout, guess not all errors should result in
			// server run to stop
			return err
		}

		// the server tracks active connections
		go func() {
			id, _ := InitConnection(ctx, conn)
			_ = id
		}()
	}
}

// InitConnection returns the client id after a successful connect and
// ack.
func InitConnection(ctx Context, conn net.Conn) (string, error) {
	var (
		sender    = tt.NewSender(conn)
		onConnect = make(chan *mq.Connect, 0)
		connwait  = tt.Intercept(onConnect)
		logger    = tt.NewLogger(tt.LevelInfo)

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
