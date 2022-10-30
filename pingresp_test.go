package mq

import (
	"fmt"
	"testing"
)

func ExamplePingResp_String() {
	p := NewPingResp()
	fmt.Println(&p)
	fmt.Print(DocumentFlags(&p))
	// output:
	// PINGRESP ---- 2 bytes
	//          3210 Size
	//
	// 3-0 reserved
}

func TestPingResp(t *testing.T) {
	p := NewPingResp()

	testControlPacket(t, &p)
	if err := p.UnmarshalBinary(nil); err != nil {
		t.Error("PingResp.UnmarshalBinary should be a noop")
	}
}
