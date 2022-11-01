package mq

import (
	"strings"
	"testing"
)

func TestUnsubscribe(t *testing.T) {
	s := NewUnsubscribe()

	eq(t, s.SetPacketID, s.PacketID, 34)

	s.AddUserProp("color", "purple")

	if v := s.String(); !strings.Contains(v, "no filters!") {
		t.Error("expect note on missing filters")
	}

	s.AddFilter("a/b/c")
	s.AddFilter("d/e")

	if v := s.String(); !strings.Contains(v, "SUBSCRIBE --1-") {
		t.Errorf("%q expect to contain %q", v, "SUBSCRIBE --1-")
	}

	testControlPacket(t, s)
}
