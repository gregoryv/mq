package mq

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func ExampleSubscribe_String() {
	s := NewSubscribe()
	s.SetSubscriptionID(39)
	s.AddUserProp("color", "purple")
	s.AddFilters(
		NewTopicFilter("a/b/c", OptQoS2|OptNL|OptRAP),
		NewTopicFilter("d/e", OptQoS1),
	)
	fmt.Println(s)
	// output:
	// SUBSCRIBE --1- p0 a/b/c --r0pn2- 37 bytes
}

func ExampleSubscribe_Dump() {
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
	// no filters
	fmt.Println(NewSubscribe())
	{ // bad qos
		s := NewSubscribe()
		s.SetSubscriptionID(1)
		s.AddFilters(NewTopicFilter("a/b", OptQoS3))
		fmt.Println(s)
	}
	{ // empty filter
		s := NewSubscribe()
		s.SetSubscriptionID(1)
		s.AddFilters(NewTopicFilter("", OptQoS1))
		fmt.Println(s)
	}
	{ // missing subscription id
		s := NewSubscribe()
		s.SetSubscriptionID(0)
		s.AddFilters(NewTopicFilter("#", OptQoS1))
		fmt.Println(s)
	}
	{ // too large subscription id
		s := NewSubscribe()
		s.SetSubscriptionID(268_435_455 + 1)
		s.AddFilters(NewTopicFilter("#", OptQoS1))
		fmt.Println(s)
	}
	// output:
	// SUBSCRIBE --1- p0  5 bytes, malformed! no filters
	// SUBSCRIBE --1- p0 a/b --r0--!! 13 bytes, malformed! invalid QoS
	// SUBSCRIBE --1- p0  --r0---1 10 bytes, malformed! empty filter
	// SUBSCRIBE --1- p0 # --r0---1 9 bytes, malformed! missing sub ID
	// SUBSCRIBE --1- p0 # --r0---1 15 bytes, malformed! too large sub ID
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
