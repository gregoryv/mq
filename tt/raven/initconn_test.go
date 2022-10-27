package raven

import (
	"bytes"
	. "context"
	"io"
	"io/ioutil"
	"testing"

	"github.com/gregoryv/mq"
)

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
}
