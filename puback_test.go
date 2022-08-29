package mqtt

import (
	"strings"
	"testing"
)

func TestPubAck(t *testing.T) {
	p := NewPubAck()
	if v := p.String(); !strings.Contains(v, "PUBACK") {
		t.Error(v)
	}

	eq(t, p.SetPacketID, p.PacketID, 99)
	// should cover the check for remaining len
	if err := testControlPacket(&p); err != nil {
		t.Error(err)
	}

	eq(t, p.SetReasonCode, p.ReasonCode, TopicNameInvalid)
	eq(t, p.SetReason, p.Reason, "name too long")

	p.AddUserProp("color", "red")

	if err := testControlPacket(&p); err != nil {
		t.Error(err)
	}

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
}
