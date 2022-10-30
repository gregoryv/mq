package mq

import (
	"fmt"
	"testing"
)

func ExampleAuth_String() {
	p := NewAuth()
	fmt.Println(&p)
	fmt.Print(DocumentFlags(&p))
	// output:
	// AUTH ---- 2 bytes
	//      3210 Size
	//
	// 3-0 reserved
}

func TestAuth(t *testing.T) {
	p := NewAuth()
	// normal disconnect
	testControlPacket(t, &p)

	eq(t, p.SetReasonCode, p.ReasonCode, MalformedPacket)
	p.AddUserProp("color", "red")
	testControlPacket(t, &p)

	// String
	if v := p.String(); v != "AUTH ---- 17 bytes" {
		t.Error(v)
	}
}
