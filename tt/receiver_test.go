package tt

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/gregoryv/mq"
)

func TestReceiver(t *testing.T) {
	{ // handler is called on packet from server
		conn, server := Dial()
		called := NewCalled()
		receiver := NewReceiver(conn, called.Handler)

		go receiver.Run(context.Background())
		server.Pub(0, "a/b", "gopher")
		<-called.Done()
	}

	{ // respects context cancellation
		conn, _ := Dial()
		receiver := NewReceiver(conn, NoopHandler)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := receiver.Run(ctx); !errors.Is(err, context.Canceled) {
			t.Errorf("unexpected error: %v", err)
		}
	}

	{ // Run is stopped on closed connection
		receiver := NewReceiver(&ClosedConn{}, NoopHandler)
		err := receiver.Run(context.Background())
		if !errors.Is(err, io.EOF) {
			t.Errorf("unexpected error: %T", err)
		}
	}
}

// ----------------------------------------

func NewCalled() *Called {
	return &Called{
		c: make(chan struct{}, 0),
	}
}

type Called struct {
	c chan struct{}
}

func (c *Called) Handler(_ context.Context, _ mq.Packet) error {
	close(c.c)
	return nil
}

func (c *Called) Done() <-chan struct{} {
	return c.c
}
