package mqtt

import (
	"fmt"
	"reflect"
	"testing"
)

func ExampleNewVarInt() {
	fmt.Println(NewVarInt(16_384))
	fmt.Println(NewVarInt(268435455))
	// output:
	// [128 128 1]
	// [255 255 255 127]
}

func TestNewVarInt(t *testing.T) {
	cases := []struct {
		x   uint
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
		if got := NewVarInt(c.x); !reflect.DeepEqual(got, c.exp) {
			t.Error("got", got, "exp", c.exp)
		}
	}
}
