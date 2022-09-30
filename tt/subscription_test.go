package tt

import (
	"strings"
	"testing"

	"github.com/gregoryv/mq"
)

func ExampleSubscription() {
	_ = NewSubscription("my/topic", func(p mq.Packet) error {
		return nil
	})

}

func TestSubscription(t *testing.T) {
	s := NewSubscription("my/topic", func(p mq.Packet) error {
		return nil
	})
	if v := s.String(); !strings.Contains(v, "my/topic") {
		t.Error(v)
	}
}
