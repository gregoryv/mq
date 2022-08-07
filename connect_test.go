package mqtt

import (
	"fmt"
	"testing"
)

func ExampleConnect_String() {
	p := NewConnect().WithFlags(0b1111_1111)
	fmt.Println(p)
	fmt.Println(p.WithFlags(0b0000_0000))
	// output:
	// CONNECT 10 MQTT5 upr2wsR
	// CONNECT 10 MQTT5 -------
}

func TestConnect(t *testing.T) {

	// WithFlags and HasFlag
	p := NewConnect().WithFlags(0b1111_0000)
	if !p.HasFlag(UsernameFlag) {
		t.Errorf("missing flag %08b", p.Flags())
	}
	if p.HasFlag(CleanStart) {
		t.Errorf("flag %08b was not set", p.Flags())
	}
}
