package mq

import (
	"strings"
	"testing"
)

func ExampleSubscription() {
	_ = NewSubscription("my/topic", func(p ControlPacket) error {
		return nil
	})
}

func TestSubscription(t *testing.T) {
	s := NewSubscription("my/topic", func(p ControlPacket) error {
		return nil
	})
	if v := s.String(); !strings.Contains(v, "my/topic") {
		t.Error(v)
	}
}
