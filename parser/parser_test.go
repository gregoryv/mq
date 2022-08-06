package parser

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/gregoryv/mqtt"
	. "github.com/gregoryv/mqtt" // for easy access to named bytes
)

func ExampleNewParser() {
	var con bytes.Buffer // some network connection
	con.Write([]byte{PUBLISH | RETAIN, 4, 0, 0, 0, 0})

	parser := NewParser(&con)
	c := make(chan *mqtt.ControlPacket, 10)
	go parser.Parse(context.Background(), c)
	pak := <-c
	fmt.Println(pak)
	// output:
	// PUBLISH-RETAIN 4
}

func TestParser(t *testing.T) {
	var con bytes.Buffer // some network connection
	con.Write([]byte{PUBLISH | RETAIN, 4, 0, 0, 0, 0})

	parser := NewParser(&con)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c := make(chan *mqtt.ControlPacket, 0) // blocks the parse loop
	err := parser.Parse(ctx, c)

	if !errors.Is(err, context.Canceled) {
		t.Error(err)
	}
}

func ExampleParseFixedHeader() {
	r := bytes.NewReader([]byte{PUBLISH | DUP, 4, 0, 0, 0, 0})
	h, _ := ParseFixedHeader(r)
	fmt.Print(h)
	// output:
	// PUBLISH-DUP 4
}

func TestParseFixedHeader(t *testing.T) {
	cases := []struct {
		in  []byte
		exp []byte
		err error
	}{
		{
			in:  []byte{PUBLISH | DUP, 4, 0, 0, 0, 0},
			exp: []byte{PUBLISH | DUP, 4},
			err: nil,
		},
		{
			in:  []byte{},
			exp: []byte{},
			err: io.EOF,
		},
		{
			in:  []byte{CONNECT},
			exp: []byte{CONNECT},
			err: io.EOF,
		},
	}
	for i, c := range cases {
		r := bytes.NewReader(c.in)
		h, err := ParseFixedHeader(r)
		if err != c.err {
			t.Fatal(i, err)
		}
		if !reflect.DeepEqual([]byte(h), c.exp) {
			t.Error("got", h, "exp", c.exp)
		}
	}
}

type reads struct {
	i        int
	sequence []interface{} // []byte or func([]byte) (int, error)
}

func (d *reads) Read(p []byte) (n int, err error) {
	if d.i >= len(d.sequence) {
		return 0, io.EOF
	}

	switch next := d.sequence[d.i].(type) {
	case []byte:
		n = copy(p, next)
	case func([]byte) (int, error):
		n, err = next(p)
	}
	d.i++
	return
}