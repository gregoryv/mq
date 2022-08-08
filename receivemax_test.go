package mqtt

import (
	"reflect"
	"testing"
)

func TestReceiveMax(t *testing.T) {
	b := ReceiveMax(76)

	data, err := b.MarshalBinary()
	if err != nil {
		t.Error("MarshalBinary", err)
	}
	if exp := []byte{0, 76}; !reflect.DeepEqual(data, exp) {
		t.Error("unexpected data ", data)
	}

	var a ReceiveMax
	if err := a.UnmarshalBinary(data); err != nil {
		t.Error("UnmarshalBinary", err)
	}

	// before and after are equal
	if b != a {
		t.Errorf("b(%v) != a(%v)", b, a)
	}
}
