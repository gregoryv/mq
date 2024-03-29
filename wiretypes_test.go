package mq

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func Test_firstByte(t *testing.T) {
	cases := []struct {
		h   byte
		exp string
	}{
		{PUBLISH | QoS2 | RETAIN, "PUBLISH -2-r"},
		{PUBLISH | QoS3, "PUBLISH -!!-"},
		{PUBLISH | DUP | QoS2, "PUBLISH d2--"},
		{PUBLISH | QoS1, "PUBLISH --1-"},
		{CONNECT, "CONNECT ----"},
	}
	for _, c := range cases {
		if got := firstByte(c.h).String(); got != c.exp {
			t.Errorf("String: %q != %q", got, c.exp)
		}
	}
}

func Test_bits(t *testing.T) {
	v := bits(0b0001_0000)
	switch {
	case !v.Has(0b0001_0000):
		t.Error("!Has")
	case v.Has(0b0000_0001):
		t.Error("Has")
	}

	var r brokenRW
	if _, err := v.ReadFrom(&r); err == nil {
		t.Error("expected error")
	}
}

func Test_wuint16(t *testing.T) {
	b := wuint16(76)

	data := make([]byte, b.width())
	b.fill(data, 0)

	if exp := []byte{0, 76}; !reflect.DeepEqual(data, exp) {
		t.Error("unexpected data ", data)
	}

	var a wuint16
	if err := a.UnmarshalBinary(data); err != nil {
		t.Error("UnmarshalBinary", err)
	}

	// before and after are equal
	if b != a {
		t.Errorf("b(%v) != a(%v)", b, a)
	}

	if got := a.width(); got != 2 {
		t.Error("invalid wuint16 width", got)
	}
}

func Test_wuint32(t *testing.T) {
	b := wuint32(76)

	data := make([]byte, b.width())
	b.fill(data, 0)

	if exp := []byte{0, 0, 0, 76}; !reflect.DeepEqual(data, exp) {
		t.Error("unexpected data ", data)
	}

	var a wuint32
	if err := a.UnmarshalBinary(data); err != nil {
		t.Error("UnmarshalBinary", err)
	}

	// before and after are equal
	if b != a {
		t.Errorf("b(%v) != a(%v)", b, a)
	}
}

func Test_vbint(t *testing.T) {
	cases := []struct {
		x   vbint
		exp []byte
	}{
		{0, []byte{0x00}},
		{127, []byte{0x7f}},

		{128, []byte{0x80, 0x01}},
		{16_383, []byte{0xff, 0x7f}},

		{148, []byte{0x94, 0x01}},

		{16_384, []byte{0x80, 0x80, 0x01}},
		{2_097_151, []byte{0xff, 0xff, 0x7f}},

		{2_097_152, []byte{0x80, 0x80, 0x80, 0x01}},
		{268_435_455, []byte{0xff, 0xff, 0xff, 0x7f}},
	}
	for _, c := range cases {
		data := make([]byte, c.x.fill(_LEN, 0))
		c.x.fill(data, 0)

		if !reflect.DeepEqual(data, c.exp) {
			t.Log("got", hex.Dump(data))
			t.Error("exp", hex.Dump(c.exp))
		}
		var after vbint
		if err := after.UnmarshalBinary(data); err != nil {
			t.Error("Unmarshal", data)
		}
		if after != c.x {
			t.Errorf("%v != %v", c.x, after)
		}

		var afterR vbint
		if _, err := afterR.ReadFrom(bytes.NewReader(data)); err != nil {
			t.Error(err)
		}
		if afterR != c.x {
			t.Errorf("%v != %v", c.x, afterR)
		}
	}

	var empty vbint
	if v := empty.fillProp(nil, 0, 0); v > 0 {
		t.Error("empty fill", v)
	}
	// error case
	var v vbint
	badData := []byte{0xff, 0xff, 0xff, 0xff, 0x7f}
	if err := v.UnmarshalBinary(badData); err == nil {
		t.Error("UnmarshalBinary should fail", badData)
	}

	if err := v.UnmarshalBinary(nil); err == nil {
		t.Error("UnmarshalBinary should fail on empty")
	}

	var r brokenRW
	if _, err := v.ReadFrom(&r); err == nil {
		t.Error("expected error")
	}

	large := []byte{0xff, 0xff, 0xff, 0xff, 0xff}
	if _, err := v.ReadFrom(bytes.NewReader(large)); err == nil {
		t.Error("expected error")
	}
}

func Test_wbool(t *testing.T) {
	var b wbool // false

	data := make([]byte, b.width())
	b.fill(data, 0)

	var a wbool
	if err := a.UnmarshalBinary(data); err != nil {
		t.Error("UnmarshalBinary", err)
	}

	// before and after are equal
	if b != a {
		t.Errorf("b(%v) != a(%v)", b, a)
	}
	// error case
	data[0] = 3
	if err := a.UnmarshalBinary(data); err == nil {
		t.Error("UnmarshalBinary should fail")
	}
}

func Test_bindata(t *testing.T) {
	indata := make([]byte, 64)
	if _, err := rand.Read(indata); err != nil {
		t.Fatal(err)
	}

	b := bindata(indata)
	data := make([]byte, b.width())
	b.fill(data, 0)

	var a bindata
	if err := a.UnmarshalBinary(data); err != nil {
		t.Error("UnmarshalBinary", err)
	}
	if err := a.UnmarshalBinary([]byte{0x00, 0x00}); err != nil {
		t.Error("UnmarshalBinary", err)
	}

	// before and after are equal
	if !reflect.DeepEqual(b, a) {
		t.Error("unmarshal -> marshal should be equal", len(b), len(a))
	}

	// error case
	if err := a.UnmarshalBinary(data[:len(data)-4]); err == nil {
		t.Error("UnmarshalBinary should fail")
	}

}

func Test_UserProp(t *testing.T) {
	b := UserProp{"key", "value"}

	data := make([]byte, b.width())
	b.fill(data, 0)

	var a UserProp
	if err := a.UnmarshalBinary(data); err != nil {
		t.Error("UnmarshalBinary", err, data)
	}

	// before and after are equal
	if !reflect.DeepEqual(b, a) {
		t.Error("unmarshal -> marshal should be equal", b, a)
	}

	if got, exp := a.String(), "key:value"; got != exp {
		t.Errorf("%q != %q", got, exp)
	}

	empty := UserProp{"", "value"} // missing key
	if v := empty.fillProp(nil, 0, 0); v > 0 {
		t.Error("empty fillProp", v)
	}
	// error cases
	if err := a.UnmarshalBinary(data[:3]); err == nil {
		t.Error("UnmarshalBinary should fail on malformed key")
	}
	if err := a.UnmarshalBinary(data[:7]); err == nil {
		t.Error("UnmarshalBinary should fail on malformed value")
	}
}

func Test_rawdata(t *testing.T) {
	defer func() {
		if e := recover(); e == nil {
			t.Error("expect panic")
		}
	}()
	var v rawdata
	v.fillProp(nil, 0, 0)
}

func Test_Ident(t *testing.T) {
	empty := Ident(9)
	if v := empty.fillProp(nil, 0, 0); v > 0 {
		t.Error("empty fillProp", v)
	}
}

var large = wstring(strings.Repeat(" ", maxUint16+1))

type brokenRW struct{}

func (w *brokenRW) Write(data []byte) (int, error) {
	return 0, broken
}

func (w *brokenRW) Read(data []byte) (int, error) {
	return 0, broken
}

var broken = fmt.Errorf("broken")
