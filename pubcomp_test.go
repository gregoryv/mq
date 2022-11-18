package mq

import (
	"fmt"
	"strings"
	"testing"
)

func ExampleNewPubComp() {
	p := NewPubComp()
	p.SetPacketID(9)
	fmt.Println(p)
	// output:
	// PUBCOMP ---- p9 Success 4 bytes
}

func TestPubComp(t *testing.T) {
	p := NewPubComp()
	if v := p.String(); !strings.Contains(v, "PUBCOMP") {
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

	// variations
	{
		p := NewPubRel()
		testControlPacket(t, p)
		if v := p.String(); !strings.Contains(v, "PUBREL") {
			t.Error(v)
		}
	}
	{
		p := NewPubRec()
		if v := p.String(); !strings.Contains(v, "PUBREC") {
			t.Error(v)
		}
	}
	{
		p := NewPubComp()
		if v := p.String(); !strings.Contains(v, "PUBCOMP") {
			t.Error(v)
		}
	}

	// type
	if a, b := NewPubComp(), NewPubRel(); a.AckType() == b.AckType() {
		t.Error("PubComp byte same as PubRel byte")
	}
}
