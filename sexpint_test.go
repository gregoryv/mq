package mqtt

import (
	"reflect"
	"testing"
	"time"
)

func TestSessionExpiryInterval(t *testing.T) {
	b := SessionExpiryInterval(76)

	data, err := b.MarshalBinary()
	if err != nil {
		t.Error("MarshalBinary", err)
	}
	if exp := []byte{0, 0, 0, 76}; !reflect.DeepEqual(data, exp) {
		t.Error("unexpected data ", data)
	}

	var a SessionExpiryInterval
	if err := a.UnmarshalBinary(data); err != nil {
		t.Error("UnmarshalBinary", err)
	}

	// before and after are equal
	if b != a {
		t.Errorf("b(%v) != a(%v)", b, a)
	}

	if got := a.String(); got != "1m16s" {
		t.Error("unexpected text", got)
	}

	if dur := a.Duration(); dur != 76*time.Second {
		t.Error("unexpected duration", dur)
	}
}
