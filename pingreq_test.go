package mq

import (
	"fmt"
	"testing"
)

func ExamplePingReq_String() {
	p := NewPingReq()
	fmt.Println(p)
	fmt.Print(DocumentFlags(p))
	// output:
	// PINGREQ ---- 2 bytes
	//         3210 Size
	//
	// 3-0 reserved
}

func TestPingReq(t *testing.T) {
	p := NewPingReq()

	testControlPacket(t, p)

	if err := p.UnmarshalBinary(nil); err != nil {
		t.Error("PingReq.UnmarshalBinary should be a noop")
	}
}
