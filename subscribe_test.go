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
	s.AddFilter("a/b/c", OptQoS2|OptNL|OptRAP)
	s.AddFilter("d/e", OptQoS1)
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

	s.AddFilter("a/b/c", OptQoS2|OptNL|OptRAP)
	s.AddFilter("d/e", OptQoS1)

	if v := s.Filters(); len(v) != 2 {
		t.Error("expect 2 filters, got", v)
	}
	if v := s.String(); !strings.Contains(v, "SUBSCRIBE --1-") {
		t.Errorf("%q expect to contain %q", v, "SUBSCRIBE --1-")
	}

	testControlPacket(t, s)
}

func TestTopicFilter(t *testing.T) {
	cases := []struct {
		f   TopicFilter
		exp string
	}{
		{
			f:   NewTopicFilter("a/b", OptQoS3),
			exp: "a/b --r0--!!",
		},
		{
			f:   NewTopicFilter("a/b", OptQoS1|OptRetain1),
			exp: "a/b --r1---1",
		},
		{
			f:   NewTopicFilter("a/b", OptQoS2|OptRetain2),
			exp: "a/b --r2--2-",
		},
		{
			f:   NewTopicFilter("a/b", OptQoS2|OptRetain3),
			exp: "a/b --!!--2-",
		},
	}

	for _, c := range cases {
		if v := c.f.String(); v != c.exp {
			t.Error("got", v, "expected", c.exp)
		}
	}
}
