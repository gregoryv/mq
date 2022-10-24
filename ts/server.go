package ts

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
// or accepting a connection fails.
func (s *Server) Run(ctx Context) error {
	l, err := net.Listen("tcp", s.bind)
	if err != nil {
		return err
	}
	defer l.Close()

	c := make(chan error, 0)

	go func() {
	loop:
		for {
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
				c <- err
			}

			// the server tracks active connections
			go func() {
				id, _ := initConnection(ctx, conn)
				_ = id
			}()
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-c:
		return err
	}
}

func initConnection(ctx Context, conn io.ReadWriteCloser) (string, error) {
	var (
		sender   = tt.NewSender(conn)
		connwait = tt.Intercept[*mq.Connect]()
		logger   = tt.NewLogger(tt.LevelInfo)

		in  = tt.NewInQueue(tt.NoopHandler, connwait, logger)
		out = tt.NewOutQueue(sender.Out, logger)
	)

	_ = out // todo register outgoing connection once connected
	go tt.NewReceiver(conn, in).Run(ctx)

	select {
	case p := <-connwait.Done():
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
