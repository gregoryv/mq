package mq

import (
	"fmt"
	"strings"
	"testing"
)

func ExamplePubAck_String() {
	p := NewPubAck()
	p.SetPacketID(9)
	fmt.Println(&p)
	fmt.Print(DocumentFlags(&p))
	// output:
	// PUBACK ---- p9 Success 4 bytes
	//        3210 PacketID Reason [reason text] Size
	//
	// 3-0 reserved
}

func ExampleNewPubComp() {
	p := NewPubComp()
	p.SetPacketID(9)
	fmt.Println(&p)
	// output:
	// PUBCOMP ---- p9 Success 4 bytes
}

func ExampleNewPubRec() {
	p := NewPubRec()
	p.SetPacketID(9)
	fmt.Println(&p)
	// output:
	// PUBREC ---- p9 Success 4 bytes
}

func ExampleNewPubRel() {
	p := NewPubRel()
	p.SetPacketID(9)
	p.SetReasonCode(PacketIdentifierNotFound)
	fmt.Println(&p)
	// output:
	// PUBREL ---- p9 PacketIdentifierNotFound 5 bytes
}

func TestPubAck(t *testing.T) {
	p := NewPubAck()
	if v := p.String(); !strings.Contains(v, "PUBACK") {
		t.Error(v)
	}

	eq(t, p.SetPacketID, p.PacketID, 99)
	// should cover the check for remaining len
	testControlPacket(t, &p)

	eq(t, p.SetReasonCode, p.ReasonCode, TopicNameInvalid)
	eq(t, p.SetReason, p.Reason, "name too long")

	p.AddUserProp("color", "red")

	testControlPacket(t, &p)

	if v := p.String(); !strings.Contains(v, "name too long") {
		t.Error(v)
	}

	// variations
	if p := NewPubRel(); !strings.Contains(p.String(), "PUBREL") {
		t.Error(p.String())
	}
	if p := NewPubRec(); !strings.Contains(p.String(), "PUBREC") {
		t.Error(p.String())
	}
	if p := NewPubComp(); !strings.Contains(p.String(), "PUBCOMP") {
		t.Error(p.String())
	}

	// type
	if a, b := NewPubAck(), NewPubRel(); a.AckType() == b.AckType() {
		t.Error("PubAck byte same as PubRel byte")
	}
}
