package mqtt

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestParse_incomplete(t *testing.T) {
	r := bytes.NewReader([]byte{CONNECT, 10, 1})
	_, err := Parse(r)
	if !errors.Is(err, ErrIncomplete) {
		t.Error("expect", ErrIncomplete, "got", err)
	}
}

func TestParse_Connect(t *testing.T) {
	p := NewConnect()
	p.SetFlags(UsernameFlag | Reserved | WillQoS1)
	r := p.Reader()
	got, err := Parse(r)
	if err != nil {
		t.Fatal(got.String(), err)
	}
	if h := got.FixedHeader(); !h.Is(CONNECT) {
		t.Error("wrong type", h)
	}
}

func TestParse_Undefined(t *testing.T) {
	r := bytes.NewReader([]byte{UNDEFINED})
	if _, err := Parse(r); err == nil {
		t.Fail()
	}
}

func TestParse_Auth(t *testing.T) {
	r := bytes.NewReader([]byte{AUTH, 0}) // AUTH is last
	// todo reverse assert once implemented, here to cover the
	// error handling of undefined
	if _, err := Parse(r); err == nil {
		t.Fail()
	}
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
			in:  []byte{0},
			exp: []byte{UNDEFINED},
			err: ErrTypeUndefined,
		},
		{
			in:  []byte{CONNECT},
			exp: []byte{CONNECT},
			err: io.EOF,
		},
	}
	for i, c := range cases {
		r := bytes.NewReader(c.in)
		h, err := parseFixedHeader(r)
		if !errors.Is(err, c.err) {
			t.Fatal(i, err)
		}
		if !reflect.DeepEqual([]byte(h), c.exp) {
			t.Error("got", []byte(h), "exp", c.exp)
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
