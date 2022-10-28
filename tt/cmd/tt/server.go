package main

import (
	. "context"
	"errors"
	"io"
	"net"
	"os"
	"time"
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

	// todo place this in a connections store
	clients map[string]io.ReadWriter
}

// Run listens for tcp connections. Blocks until context is cancelled
// or accepting a connection fails. Accepting new connection can only
// be interrupted if listener has SetDeadline method.
func (s *Server) Run(l net.Listener, ctx Context) error {
loop:
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		// timeout Accept call so we don't block the loop
		if l, ok := l.(interface{ SetDeadline(time.Time) error }); ok {
			l.SetDeadline(time.Now().Add(s.acceptTimeout))
		}
		conn, err := l.Accept()
		if errors.Is(err, os.ErrDeadlineExceeded) {
			continue loop
		}

		if err != nil {
			// todo check what causes Accept to fail other than
			// timeout, guess not all errors should result in
			// server run to stop
			return err
		}

		// the server tracks active connections
		go func() {
			id, _ := InitConn(ctx, conn)
			_ = id
		}()
	}
}
