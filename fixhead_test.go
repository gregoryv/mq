package mqtt

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"
)

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

func ExampleFixedHeader_String() {
	fmt.Println(new(FixedHeader).String())
	fmt.Println(FixedHeader{PUBLISH})
	fmt.Println(FixedHeader{PUBLISH | DUP})
	fmt.Println(FixedHeader{PUBLISH | DUP | RETAIN})
	fmt.Println(FixedHeader{PUBLISH | QoS2, 2})
	//output:
	// FORBIDDEN
	// PUBLISH
	// PUBLISH-DUP
	// PUBLISH-DUP-RETAIN
	// PUBLISH-QoS2 2
}

func ExampleFixedHeader_HasFlag() {
	a := FixedHeader{DUP}
	fmt.Printf("%08b %v\n", a[0], a.HasFlag(DUP))
	b := FixedHeader{0x00}
	fmt.Printf("%08b %v\n", b[0], b.HasFlag(DUP))
	// output:
	// 00001000 true
	// 00000000 false
}

func ExampleFixedHeader_Name() {
	all := []FixedHeader{
		FixedHeader{FORBIDDEN},
		FixedHeader{CONNECT},
		FixedHeader{CONNACK},
		FixedHeader{PUBLISH},
		FixedHeader{PUBACK},
		FixedHeader{PUBREC},
		FixedHeader{PUBREL},
		FixedHeader{PUBCOMP},
		FixedHeader{SUBSCRIBE},
		FixedHeader{SUBACK},
		FixedHeader{UNSUBSCRIBE},
		FixedHeader{UNSUBACK},
		FixedHeader{PINGREQ},
		FixedHeader{PINGRESP},
		FixedHeader{DISCONNECT},
		FixedHeader{AUTH},
	}
	for _, h := range all {
		fmt.Printf("%08b 0x%02x %s\n", h.Value(), h.Value(), h.Name())
	}
	// output:
	// 00000000 0x00 FORBIDDEN
	// 00010000 0x10 CONNECT
	// 00100000 0x20 CONNACK
	// 00110000 0x30 PUBLISH
	// 01000000 0x40 PUBACK
	// 01010000 0x50 PUBREC
	// 01100000 0x60 PUBREL
	// 01110000 0x70 PUBCOMP
	// 10000000 0x80 SUBSCRIBE
	// 10010000 0x90 SUBACK
	// 10100000 0xa0 UNSUBSCRIBE
	// 10110000 0xb0 UNSUBACK
	// 11000000 0xc0 PINGREQ
	// 11010000 0xd0 PINGRESP
	// 11100000 0xe0 DISCONNECT
	// 11110000 0xf0 AUTH
}
