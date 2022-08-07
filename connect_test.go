package mqtt

import (
	"fmt"
	"testing"
	"time"
)

func ExampleConnect() {
	p := NewConnect().WithFlags(0b1111_1111)
	fmt.Println(p)

	p.WithFlags(0b0000_0000)
	p.SetSessionExpiryInterval(132 * time.Second)
	fmt.Println(p)
	// output:
	// CONNECT 15 MQTT5 upr2wsR 59s
	// CONNECT 15 MQTT5 ------- 2m12s
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
