package mqtt

import (
	"crypto/rand"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------
// Headers
// ---------------------------------------------------------------------

func TestfirstByte(t *testing.T) {
	cases := []struct {
		h   firstByte
		exp string
	}{
		{firstByte(PUBLISH | QoS2 | RETAIN), "PUBLISH -2-r"},
		{firstByte(PUBLISH | QoS1 | QoS2), "PUBLISH -!!-"},
		{firstByte(PUBLISH | DUP | QoS2), "PUBLISH d2--"},
		{firstByte(PUBLISH | QoS1), "PUBLISH --1-"},
		{firstByte(CONNECT), "CONNECT ----"},
	}
	for _, c := range cases {
		if got, exp := c.h.String(), c.exp; got != exp {
			t.Errorf("String: %q != %q", got, exp)
		}
	}
}

// ---------------------------------------------------------------------
// Data representations, the low level data types
// ---------------------------------------------------------------------

func Test_Bits(t *testing.T) {
	v := Bits(0b0001_0000)
	switch {
	case !v.Has(0b0001_0000):
		t.Error("!Has")
	case v.Has(0b0000_0001):
		t.Error("Has")
	}

}

func Test_b2int(t *testing.T) {
	b := b2int(76)

	data := make([]byte, b.width())
	b.fill(data, 0)

	if exp := []byte{0, 76}; !reflect.DeepEqual(data, exp) {
		t.Error("unexpected data ", data)
	}

	var a b2int
	if err := a.UnmarshalBinary(data); err != nil {
		t.Error("UnmarshalBinary", err)
	}

	// before and after are equal
	if b != a {
		t.Errorf("b(%v) != a(%v)", b, a)
	}

	if got := a.width(); got != 2 {
		t.Error("invalid b2int width", got)
	}
}

func Test_b4int(t *testing.T) {
	b := b4int(76)

	data := make([]byte, b.width())
	b.fill(data, 0)

	if exp := []byte{0, 0, 0, 76}; !reflect.DeepEqual(data, exp) {
		t.Error("unexpected data ", data)
	}

	var a b4int
	if err := a.UnmarshalBinary(data); err != nil {
		t.Error("UnmarshalBinary", err)
	}

	// before and after are equal
	if b != a {
		t.Errorf("b(%v) != a(%v)", b, a)
	}
}

// ................................................ Data representations

func Test_u8str(t *testing.T) {
	b := u8str("۞ gopher från sverige")

	data := make([]byte, b.width())
	b.fill(data, 0)

	var a u8str
	if err := a.UnmarshalBinary(data); err != nil {
		t.Error("UnmarshalBinary", err)
	}

	// before and after are equal
	if b != a {
		t.Errorf("b(%v) != a(%v)", b, a)
	}

	// error case
	if err := a.UnmarshalBinary(data[:len(data)-4]); err == nil {
		t.Error("UnmarshalBinary should fail")
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

		{16_384, []byte{0x80, 0x80, 0x01}},
		{2_097_151, []byte{0xff, 0xff, 0x7f}},

		{2_097_152, []byte{0x80, 0x80, 0x80, 0x01}},
		{268_435_455, []byte{0xff, 0xff, 0xff, 0x7f}},
	}
	for _, c := range cases {
		data := make([]byte, c.x.width())
		c.x.fill(data, 0)

		if !reflect.DeepEqual(data, c.exp) {
			t.Error("got", data, "exp", c.exp)
		}
		var after vbint
		if err := after.UnmarshalBinary(data); err != nil {
			t.Error("Unmarshal", data)
		}
		if after != c.x {
			t.Errorf("%v != %v", c.x, after)
		}
		// widths
		if got := c.x.width(); got != len(c.exp) {
			t.Error("unexpected width", got, c.x)
		}
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

}

// ................................................ Data representations

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

	// before and after are equal
	if !reflect.DeepEqual(b, a) {
		t.Error("unmarshal -> marshal should be equal", len(b), len(a))
	}

	// error case
	if err := a.UnmarshalBinary(data[:len(data)-4]); err == nil {
		t.Error("UnmarshalBinary should fail")
	}

}

// ................................................ Data representations

func Test_property(t *testing.T) {
	b := property{"key", "value"}

	data := make([]byte, b.width())
	b.fill(data, 0)

	var a property
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

	// error cases
	if err := a.UnmarshalBinary(data[:3]); err == nil {
		t.Error("UnmarshalBinary should fail on malformed key")
	}
	if err := a.UnmarshalBinary(data[:7]); err == nil {
		t.Error("UnmarshalBinary should fail on malformed value")
	}
}

var large = u8str(strings.Repeat(" ", MaxUint16+1))

// ................................................ Data representations

func ExampleMalformed_Error() {
	e := Malformed{
		method: "unmarshal",
		t:      fmt.Sprintf("%T", Connect{}),
		reason: "missing data",
	}
	fmt.Println(e.Error())
	e.ref = "remaining length"
	fmt.Println(e.Error())
	// output:
	// malformed mqtt.Connect unmarshal: missing data
	// malformed mqtt.Connect unmarshal: remaining length missing data
}

type brokenWriter struct{}

func (w *brokenWriter) Write(data []byte) (int, error) {
	return 0, broken
}

var broken = fmt.Errorf("broken")
