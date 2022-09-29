package mq

import (
	"reflect"
	"testing"
)

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

	testControlPacket(t, &p)
}
