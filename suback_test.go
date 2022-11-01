package mq

import (
	"fmt"
	"reflect"
	"testing"
)

func ExampleSubAck_String() {
	p := NewSubAck()
	p.SetPacketID(3)
	fmt.Println(p)
	fmt.Print(DocumentFlags(p))
	// output:
	// SUBACK ---- p3 5 bytes
	//        3210 PacketID Size
	//
	// 3-0 reserved
}

func TestSubAck(t *testing.T) {
	p := NewSubAck()

	eq(t, p.SetPacketID, p.PacketID, 99)
	eq(t, p.SetReasonString, p.ReasonString, "gopher")

	p.AddUserProp("color", "red")
	p.AddReasonCode(GrantedQoS0)
	p.AddReasonCode(GrantedQoS1)

	if v := p.ReasonCodes(); !reflect.DeepEqual(v, []byte{0x00, 0x01}) {
		t.Error(v)
	}

	testControlPacket(t, p)
}
