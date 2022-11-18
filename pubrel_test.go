package mq

import (
	"fmt"
	"strings"
	"testing"
)

func ExampleNewPubRel() {
	p := NewPubRel()
	p.SetPacketID(9)
	p.SetReasonCode(PacketIdentifierNotFound)
	fmt.Println(p)
	// output:
	// PUBREL --1- p9 PacketIdentifierNotFound 5 bytes
}

func TestPubRel(t *testing.T) {
	p := NewPubRel()
	if v := p.String(); !strings.Contains(v, "PUBREL") {
		t.Error(v)
	}

	eq(t, p.SetPacketID, p.PacketID, 99)
	// should cover the check for remaining len
	testControlPacket(t, p)

	eq(t, p.SetReasonCode, p.ReasonCode, TopicNameInvalid)
	eq(t, p.SetReason, p.Reason, "name too long")

	p.AddUserProp("color", "red")

	testControlPacket(t, p)

	if v := p.String(); !strings.Contains(v, "name too long") {
		t.Error(v)
	}
}
