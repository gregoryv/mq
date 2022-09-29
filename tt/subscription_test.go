package tt

import (
	"testing"

	"github.com/gregoryv/mq"
)

func TestSubscription(t *testing.T) {
	s := NewSubscription()

	p := mq.NewSubscribe()
	// configure settings...

	s.SetPacket(&p)
	s.SetHandler(ignore)

	t.Error("continue here")
}
