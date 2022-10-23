package pink

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"time"

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

	acceptTimeout  time.Duration
	connectTimeout time.Duration // client has to send the initial connect packet

	io.Writer

	clients map[string]io.ReadWriter
}

func (s *Server) Run(ctx context.Context) error {
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
				c <- err
			}

			// the server tracks active connections
			go s.AddConnection(ctx, conn)
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-c:
		return err
	}
}

func (s *Server) AddConnection(ctx context.Context, conn io.ReadWriter) {
	// todo create in/out queues for each connection
	var (
		sender = tt.NewSender(conn)
		logger = tt.NewLogger(tt.LevelInfo)

		in  = tt.NewInQueue(tt.NoopHandler, logger)
		out = tt.NewOutQueue(sender.Out, logger)
	)
	_ = out // todo register outgoing connection once connected
	go tt.NewReceiver(conn, in).Run(ctx)
	s.Writer = conn
}
