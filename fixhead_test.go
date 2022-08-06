package mqtt

import (
	"fmt"
	"testing"
)

func ExampleFixedHeader_String() {
	fmt.Println(new(FixedHeader).String())
	fmt.Println(FixedHeader{PUBLISH, 39})
	fmt.Println(FixedHeader{PUBLISH | DUP | RETAIN})
	fmt.Println(FixedHeader{PUBLISH | QoS2, 2})
	//output:
	// UNDEFINED
	// PUBLISH 39
	// PUBLISH-DUP-RETAIN
	// PUBLISH-QoS2 2
}

func TestFixedHeader(t *testing.T) {
	h := FixedHeader([]byte{PUBLISH | DUP})

	if h.Is(CONNECT) {
		t.Error("!Is", CONNECT)
	}
	if h.HasFlag(RETAIN) {
		t.Error("!HasFlag", RETAIN)
	}
}
