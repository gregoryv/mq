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

func ExampleFixedHeader() {
	bad := []byte{PUBLISH | QoS1 | QoS2}
	var f FixedHeader
	fmt.Println(f.UnmarshalBinary(bad))
	// output:
	// malformed mqtt.FixedHeader unmarshal: remaining length missing data
}

func TestFixedHeader(t *testing.T) {
	b := FixedHeader{
		header:       PUBLISH | DUP | QoS1,
		remainingLen: 10,
	}
	// marshaling
	data, err := b.MarshalBinary()
	if err != nil {
		t.Error("MarshalBinary", err)
	}

	var a FixedHeader
	if err := a.UnmarshalBinary(data); err != nil {
		t.Error("UnmarshalBinary", err)
	}

	// other methods
	if a.Is(CONNECT) {
		t.Error("!Is", CONNECT)
	}
	if a.HasFlag(RETAIN) {
		t.Error("!HasFlag", RETAIN)
	}
	cases := []struct {
		h   FixedHeader
		exp string
	}{
		{
			h:   a,
			exp: "PUBLISH d-1- 10",
		},
		{
			h:   FixedHeader{header: PUBLISH | QoS2 | RETAIN},
			exp: "PUBLISH -2-r 0",
		},
		{
			h:   FixedHeader{header: PUBLISH | QoS1 | QoS2},
			exp: "PUBLISH -!!- 0",
		},
		{
			h:   FixedHeader{header: CONNECT},
			exp: "CONNECT ---- 0",
		},
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

func Test_bits(t *testing.T) {
	v := bits(0b0001_0000)
	switch {
	case !v.Has(0b0001_0000):
		t.Error("!Has")
	case v.Has(0b0000_0001):
		t.Error("Has")
	}

}

func Test_b2int(t *testing.T) {
	b := b2int(76)

	data, err := b.MarshalBinary()
	if err != nil {
		t.Error("MarshalBinary", err)
	}
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
}

func Test_b4int(t *testing.T) {
	b := b4int(76)

	data, err := b.MarshalBinary()
	if err != nil {
		t.Error("MarshalBinary", err)
	}
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

	data, err := b.MarshalBinary()
	if err != nil {
		t.Error("MarshalBinary", err)
	}

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

	large := strings.Repeat(" ", MaxUint16+1)
	if _, err := u8str(large).MarshalBinary(); err == nil {
		t.Error("MarshalBinary should fail when len > MaxUint16")
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
		data, _ := c.x.MarshalBinary()
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
		if got := c.x.Width(); got != len(c.exp) {
			t.Error("unexpected width", got, c.x)
		}
	}

	// error case
	var v vbint
	badData := []byte{0xff, 0xff, 0xff, 0xff, 0x7f}
	if err := v.UnmarshalBinary(badData); err == nil {
		t.Error("UnmarshalBinary should fail", badData)
	}

}

// ................................................ Data representations

func Test_bindat(t *testing.T) {
	indata := make([]byte, 64)
	if _, err := rand.Read(indata); err != nil {
		t.Fatal(err)
	}

	b := bindat(indata)
	data, err := b.MarshalBinary()
	if err != nil {
		t.Error("MarshalBinary", err)
	}

	var a bindat
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

	large := make([]byte, MaxUint16+1)
	if _, err = rand.Read(large); err != nil {
		t.Fatal(err)
	}
	if _, err := bindat(large).MarshalBinary(); err == nil {
		t.Error("MarshalBinary should fail when len > MaxUint16")
	}
}

// ................................................ Data representations

func Test_spair(t *testing.T) {
	b := spair{"key", "value"}

	data, err := b.MarshalBinary()
	if err != nil {
		t.Error("MarshalBinary", err)
	}

	var a spair
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

	// large key
	large := u8str(strings.Repeat(" ", MaxUint16+1))
	c := spair{large, ""}
	if _, err := c.MarshalBinary(); err == nil {
		t.Error("MarshalBinary should fail on malformed key")
	}
	// large value
	d := spair{"key", large}
	if _, err := d.MarshalBinary(); err == nil {
		t.Error("MarshalBinary should fail on malformed value")
	}
}

var large = u8str(strings.Repeat(" ", MaxUint16+1))

// ................................................ Data representations
