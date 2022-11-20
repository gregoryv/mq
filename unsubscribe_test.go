package mq

import (
	"os"
	"strings"
	"testing"
)

func ExampleDump_unsubscribe() {
	s := NewUnsubscribe()
	s.AddUserProp("color", "purple")
	s.AddFilter("a/b/c")
	s.AddFilter("d/e")
	Dump(os.Stdout, s)
	// output:
	// PacketID: 0
	// Filters
	//   0. a/b/c
	//   1. d/e
	// UserProperties
	//   0. color: "purple"
}

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
