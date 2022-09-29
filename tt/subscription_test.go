package tt

import (
	"context"
	"fmt"
	"testing"

	"github.com/gregoryv/mq"
)

func TestSubscription(t *testing.T) {
	s := NewSubscription()

	p := mq.NewSubscribe()
	// configure settings...

	s.SetPacket(&p)
	s.SetHandler(func(_ context.Context, p mq.Packet) error {
		return fmt.Errorf(": todo")
	})
}
