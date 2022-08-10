package mqtt

import (
	"crypto/rand"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------
// 3.1.2.3 Connect Flags
// ---------------------------------------------------------------------

func TestConnectFlags(t *testing.T) {
	f := ConnectFlags(0b11110110)
	// QoS2
	if got, exp := f.String(), "upr2ws-"; got != exp {
		t.Errorf("got %q != exp %q", got, exp)
	}
	// QoS1
	f = ConnectFlags(0b11101110)
	if got, exp := f.String(), "upr1ws-"; got != exp {
		t.Errorf("got %q != exp %q", got, exp)
	}
	f = ConnectFlags(0b00000001)
	if got, exp := f.String(), "------!"; got != exp {
		t.Errorf("got %q != exp %q", got, exp)
	}
	if f.Has(WillFlag) || !f.Has(Reserved) {
		t.Errorf("Has %08b", f)
	}
}

// ---------------------------------------------------------------------
// 3.1.2.11 CONNECT Properties
// ---------------------------------------------------------------------

func TestConnectProperties(t *testing.T) {
	b := SessionExpiryInterval(76)

	if got := b.String(); got != "1m16s" {
		t.Error("unexpected text", got)
	}

	if dur := b.Duration(); dur != 76*time.Second {
		t.Error("unexpected duration", dur)
	}
}

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

func TestBits(t *testing.T) {
	v := Bits(0b0001_0000)
	switch {
	case !v.Has(0b0001_0000):
		t.Error("!Has")
	case v.Has(0b0000_0001):
		t.Error("Has")
	}

}

func TestTwoByteInt(t *testing.T) {
	b := TwoByteInt(76)

	data, err := b.MarshalBinary()
	if err != nil {
		t.Error("MarshalBinary", err)
	}
	if exp := []byte{0, 76}; !reflect.DeepEqual(data, exp) {
		t.Error("unexpected data ", data)
	}

	var a TwoByteInt
	if err := a.UnmarshalBinary(data); err != nil {
		t.Error("UnmarshalBinary", err)
	}

	// before and after are equal
	if b != a {
		t.Errorf("b(%v) != a(%v)", b, a)
	}
}

func TestFourByteInt(t *testing.T) {
	b := FourByteInt(76)

	data, err := b.MarshalBinary()
	if err != nil {
		t.Error("MarshalBinary", err)
	}
	if exp := []byte{0, 0, 0, 76}; !reflect.DeepEqual(data, exp) {
		t.Error("unexpected data ", data)
	}

	var a FourByteInt
	if err := a.UnmarshalBinary(data); err != nil {
		t.Error("UnmarshalBinary", err)
	}

	// before and after are equal
	if b != a {
		t.Errorf("b(%v) != a(%v)", b, a)
	}
}

// ................................................ Data representations

func TestUTF8String(t *testing.T) {
	b := UTF8String("۞ gopher från sverige")

	data, err := b.MarshalBinary()
	if err != nil {
		t.Error("MarshalBinary", err)
	}

	var a UTF8String
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
	if _, err := UTF8String(large).MarshalBinary(); err == nil {
		t.Error("MarshalBinary should fail when len > MaxUint16")
	}
}

func TestVarByteInt(t *testing.T) {
	cases := []struct {
		x   VarByteInt
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
		var after VarByteInt
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
	var v VarByteInt
	badData := []byte{0xff, 0xff, 0xff, 0xff, 0x7f}
	if err := v.UnmarshalBinary(badData); err == nil {
		t.Error("UnmarshalBinary should fail", badData)
	}
}

// ................................................ Data representations

func TestBinaryData(t *testing.T) {
	indata := make([]byte, 64)
	if _, err := rand.Read(indata); err != nil {
		t.Fatal(err)
	}

	b := BinaryData(indata)
	data, err := b.MarshalBinary()
	if err != nil {
		t.Error("MarshalBinary", err)
	}

	var a BinaryData
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
	if _, err := BinaryData(large).MarshalBinary(); err == nil {
		t.Error("MarshalBinary should fail when len > MaxUint16")
	}
}

// ................................................ Data representations

func TestUTF8StringPair(t *testing.T) {
	b := UTF8StringPair{"key", "value"}

	data, err := b.MarshalBinary()
	if err != nil {
		t.Error("MarshalBinary", err)
	}

	var a UTF8StringPair
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
	large := UTF8String(strings.Repeat(" ", MaxUint16+1))
	c := UTF8StringPair{large, ""}
	if _, err := c.MarshalBinary(); err == nil {
		t.Error("MarshalBinary should fail on malformed key")
	}
	// large value
	d := UTF8StringPair{"key", large}
	if _, err := d.MarshalBinary(); err == nil {
		t.Error("MarshalBinary should fail on malformed value")
	}
}

func ExampleUTF8StringPair() {
	_, err := (&UTF8StringPair{large, ""}).MarshalBinary()
	fmt.Println(err)
	// output:
	// malformed mqtt.UTF8StringPair marshal: key size exceeded
}

func ExampleVarByteInt() {
	badData := []byte{0xff, 0xff, 0xff, 0xff, 0x7f}
	fmt.Println(new(VarByteInt).UnmarshalBinary(badData))
	// output:
	// malformed mqtt.VarByteInt unmarshal: size exceeded
}

// ................................................ Data representations

func ExampleBinaryData() {
	_, err := BinaryData(large).MarshalBinary()
	fmt.Println(err)

	var bin BinaryData
	fmt.Println(bin.UnmarshalBinary([]byte{0, 2}))
	// output:
	// malformed mqtt.BinaryData marshal: size exceeded
	// malformed mqtt.BinaryData unmarshal: missing data
}

func ExampleUTF8String() {
	_, err := UTF8String(large).MarshalBinary()
	fmt.Println(err)

	var s UTF8String
	fmt.Println(s.UnmarshalBinary([]byte{0, 2}))
	// output:
	// malformed mqtt.UTF8String marshal: size exceeded
	// malformed mqtt.UTF8String unmarshal: missing data
}

var large = UTF8String(strings.Repeat(" ", MaxUint16+1))

// ................................................ Data representations
