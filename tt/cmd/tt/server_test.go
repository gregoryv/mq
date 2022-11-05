package main

import (
	"bytes"
	. "context"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"testing"
	"time"

	"github.com/gregoryv/mq"
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

// gomerge src: initconn_test.go

func TestInitConn(t *testing.T) {
	fromClient, server := io.Pipe()
	defer server.Close()
	toClient := ioutil.Discard
	conn := NewConn(fromClient, toClient)

	{ // connect
		p := mq.NewConnect()
		p.SetClientID("test-id")
		go p.WriteTo(server)
	}

	var logs bytes.Buffer
	NextLogWriter = &logs

	id, err := InitConn(Background(), conn)
	if err != nil {
		t.Fatal(err)
	}
	if id != "test-id" {
		t.Log(logs.String())
		t.Error("got", id)
	}

	// todo respects cancel

	// todo does not leek receiver run

	// todo decide if InitConn should only be running during the
	// connection and once ok, switch to another
}
