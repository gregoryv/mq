package raven

import (
	. "context"
	"errors"
	"io"
	"net"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	s := NewServer()

	// Accept respects deadline
	ctx, cancel := WithCancel(Background())
	time.AfterFunc(2*s.acceptTimeout, cancel)

	l, err := net.Listen("tcp", ":")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Run(l, ctx); !errors.Is(err, Canceled) {
		t.Error(err)
	}

	// Ends on listener close
	time.AfterFunc(time.Millisecond, func() { l.Close() })
	if err := s.Run(l, Background()); !errors.Is(err, net.ErrClosed) {
		t.Error(err)
	}
}

// ----------------------------------------

func NewConn(r io.Reader, w io.Writer) *Conn {
	return &Conn{Reader: r, Writer: w}
}

type Conn struct {
	io.Reader // incoming from server
	io.Writer // outgoing to server
}

func (c *Conn) Close() error {
	if v, ok := c.Reader.(io.Closer); ok {
		if err := v.Close(); err != nil {
			return err
		}
	}
	if v, ok := c.Writer.(io.Closer); ok {
		if err := v.Close(); err != nil {
			return err
		}
	}
	return nil
}