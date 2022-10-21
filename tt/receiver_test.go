package tt

import (
	"context"
	"errors"
	"io"
	"net"
	"testing"
	"time"

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
		// create a tcp server
		ln, err := net.Listen("tcp", ":")
		if err != nil {
			t.Fatal(err)
		}
		defer ln.Close()
		// connect to it
		conn, err := net.Dial("tcp", ln.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		receiver := NewReceiver(conn, NoopHandler)
		receiver.readTimeout = time.Microsecond // speedup

		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(2*time.Microsecond, cancel)
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
