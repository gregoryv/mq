package tt

import (
	"context"
	"fmt"
	"testing"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/proto"
)

func TestSubscription(t *testing.T) {
	s := NewSubscription()

	p := mq.NewSubscribe()
	// configure settings...

	s.SetPacket(&p)
	s.SetHandler(func(_ context.Context, p proto.Packet) error {
		return fmt.Errorf(": todo")
	})
}
