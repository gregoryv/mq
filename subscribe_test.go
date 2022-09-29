package mq

import (
	"strings"
	"testing"
)

func TestSubscribe(t *testing.T) {
	s := NewSubscribe()

	eq(t, s.SetPacketID, s.PacketID, 34)
	eq(t, s.SetSubscriptionID, s.SubscriptionID, 99)

	s.AddUserProp("color", "purple")

	if v := s.String(); !strings.Contains(v, "no filters!") {
		t.Error("expect note on missing filters")
	}

	s.AddFilter("a/b/c", FopQoS2|FopNL|FopRAP)
	s.AddFilter("d/e", FopQoS1)

	if v := s.String(); !strings.Contains(v, "SUBSCRIBE --1-") {
		t.Errorf("%q expect to contain %q", v, "SUBSCRIBE --1-")
	}

	testControlPacket(t, &s)
}

func TestTopicFilter(t *testing.T) {
	cases := []struct {
		f   TopicFilter
		exp string
	}{
		{
			f:   NewTopicFilter("a/b", FopQoS3),
			exp: "a/b --r0--!!",
		},
		{
			f:   NewTopicFilter("a/b", FopQoS1|FopRetain1),
			exp: "a/b --r1---1",
		},
		{
			f:   NewTopicFilter("a/b", FopQoS2|FopRetain2),
			exp: "a/b --r2--2-",
		},
		{
			f:   NewTopicFilter("a/b", FopQoS2|FopRetain3),
			exp: "a/b --!!--2-",
		},
	}

	for _, c := range cases {
		if v := c.f.String(); v != c.exp {
			t.Error("got", v, "expected", c.exp)
		}
	}
}
