package main

import (
	. "context"
	"errors"
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
