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
	// run in background
	l, err := net.Listen("tcp", ":")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := WithCancel(Background())
	go func() {
		err = s.Run(l, ctx)
	}()

	// Accept connection
	conn, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	conn.Close()

	// Accept respects deadline
	cancel()
	<-time.After(2 * s.acceptTimeout)
	if !errors.Is(err, Canceled) {
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
