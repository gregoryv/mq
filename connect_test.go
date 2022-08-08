package mqtt

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func ExampleConnect() {
	p := NewConnect().WithFlags(0b1111_1111)
	fmt.Println(p)

	p.WithFlags(0b0000_0000)
	p.SessionExpiryInterval = 132
	fmt.Println(p)
	// output:
	// CONNECT 15 MQTT5 upr2wsR 59s
	// CONNECT 15 MQTT5 ------- 2m12s
}

func TestConnect_MarshalBinary(t *testing.T) {
	if _, err := NewConnect().MarshalBinary(); err != nil {
		t.Fatal(err)
	}
}

func xTestParse_Connect(t *testing.T) {
	p := NewConnect()
	p.SetFlags(UsernameFlag | Reserved | WillQoS1)

	data, _ := p.MarshalBinary()
	got := mustParse(t, bytes.NewReader(data)).(*Connect)
	if !reflect.DeepEqual(p, got) {
		t.Log("exp", p)
		t.Error("got", got)
	}
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
