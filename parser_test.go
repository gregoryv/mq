package mqtt

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func Example() {
	var (
		payload = []byte("interesting coding")
		stream  = append([]byte{
			PUBLISH | RETAIN, byte(len(payload)),
		}, payload...)
		parser = NewParser(bytes.NewReader(stream))
		c      = make(chan *ControlPacket, 0)
	)
	go parser.Parse(context.Background(), c)
	fmt.Println(<-c)
	// output:
	// PUBLISH-RETAIN 18 "interesting coding"
}

func ExampleNewParser() {
	var (
		con    = bytes.NewReader([]byte{PUBLISH | RETAIN, 4, 0, 0, 0, 0})
		parser = NewParser(con)
		c      = make(chan *ControlPacket, 10)
	)
	go parser.Parse(context.Background(), c)
	fmt.Println(<-c)
	// output:
	// PUBLISH-RETAIN 4
}

func TestParser_respectsContextCancel(t *testing.T) {
	var (
		con         = bytes.NewReader([]byte{PUBLISH | RETAIN, 4, 0, 0, 0, 0})
		parser      = NewParser(con)
		c           = make(chan *ControlPacket, 0) // blocks the parse loop
		ctx, cancel = context.WithCancel(context.Background())
	)
	cancel()
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
	SetOutput(os.Stderr)
	defer SetOutput(ioutil.Discard)

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
		if !errors.Is(err, c.err) {
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
