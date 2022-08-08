package mqtt

import (
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
	}
	// error case
	var v VarByteInt
	badData := []byte{0xff, 0xff, 0xff, 0xff, 0x7f}
	if err := v.UnmarshalBinary(badData); err == nil {
		t.Error("UnmarshalBinary should fail", badData)
	}
}
