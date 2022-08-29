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
}
