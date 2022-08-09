package mqtt

import (
	"fmt"
	"testing"
)

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

func ExampleFixedHeader() {
	bad := []byte{PUBLISH | QoS1 | QoS2}
	var f FixedHeader
	fmt.Println(f.UnmarshalBinary(bad))
	// output:
	// malformed mqtt.FixedHeader unmarshal: remaining length missing data
}
