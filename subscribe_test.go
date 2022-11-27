package mq

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func ExampleSubscribe() {
	s := NewSubscribe()
	s.AddUserProp("color", "purple")
	s.AddFilters(
		NewTopicFilter("a/b/c", OptQoS2|OptNL|OptRAP),
		NewTopicFilter("d/e", OptQoS1),
	)
	Dump(os.Stdout, s)
	// output:
	// PacketID: 0
	// SubscriptionID: 0
	// Filters
	//   0. a/b/c --r0pn2-
	//   1. d/e --r0---1
	// UserProperties
	//   0. color: "purple"
}

func ExampleSubscribe_malformed() {
	s := NewSubscribe()
	fmt.Println(s)
	// output:
	// SUBSCRIBE --1- p0  5 bytes, malformed! no filters
}

func TestSubscribe(t *testing.T) {
	s := NewSubscribe()

	eq(t, s.SetPacketID, s.PacketID, 34)
	eq(t, s.SetSubscriptionID, s.SubscriptionID, 99)

	s.AddUserProp("color", "purple")

	if v := s.String(); !strings.Contains(v, "malformed! no filters") {
		t.Error("expect note on missing filters")
	}

	s.AddFilters(
		NewTopicFilter("a/b/c", OptQoS2|OptNL|OptRAP),
		NewTopicFilter("d/e", OptQoS1),
	)

	if v := s.Filters(); len(v) != 2 {
		t.Error("expect 2 filters, got", v)
	}
	if v := s.String(); !strings.Contains(v, "SUBSCRIBE --1-") {
		t.Errorf("%q expect to contain %q", v, "SUBSCRIBE --1-")
	}

	testControlPacket(t, s)
}
