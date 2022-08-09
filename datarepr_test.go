package mqtt

import (
	"crypto/rand"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

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

func ExampleMalformed() {
	_, err := (&UTF8StringPair{large, ""}).MarshalBinary()
	fmt.Println(err)

	badData := []byte{0xff, 0xff, 0xff, 0xff, 0x7f}
	fmt.Println(new(VarByteInt).UnmarshalBinary(badData))

	_, err = BinaryData(large).MarshalBinary()
	fmt.Println(err)

	var bin BinaryData
	fmt.Println(bin.UnmarshalBinary([]byte{0, 2}))

	_, err = UTF8String(large).MarshalBinary()
	fmt.Println(err)

	var s UTF8String
	fmt.Println(s.UnmarshalBinary([]byte{0, 2}))
	// output:
	// malformed mqtt.UTF8StringPair marshal: key size exceeded
	// malformed mqtt.VarByteInt unmarshal: size exceeded
	// malformed mqtt.BinaryData marshal: size exceeded
	// malformed mqtt.BinaryData unmarshal: missing data
	// malformed mqtt.UTF8String marshal: size exceeded
	// malformed mqtt.UTF8String unmarshal: missing data
}

var large = UTF8String(strings.Repeat(" ", MaxUint16+1))
