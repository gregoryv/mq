package mqtt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func Example() {
	got, _ := Parse(NewConnect().Reader())
	fmt.Println(got.FixedHeader())
	// output:
	// CONNECT 6
}

func TestParse_Undefined(t *testing.T) {
	r := bytes.NewReader([]byte{UNDEFINED})
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
