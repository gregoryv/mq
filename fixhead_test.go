package mqtt

import (
	"testing"
)

func TestFixedHeader(t *testing.T) {
	h := FixedHeader{
		header:  PUBLISH | DUP | QoS1,
		content: []byte("gopher"),
	}

	if h.Is(CONNECT) {
		t.Error("!Is", CONNECT)
	}
	if h.HasFlag(RETAIN) {
		t.Error("!HasFlag", RETAIN)
	}

	cases := []struct {
		h   FixedHeader
		exp string
	}{
		{
			h:   h,
			exp: "PUBLISH d-1- 6",
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
