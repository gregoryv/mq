package x

import (
	"context"
	"fmt"
	"testing"

	"github.com/gregoryv/mqtt"
	"github.com/gregoryv/mqtt/proto"
)

func TestSubscription(t *testing.T) {
	s := NewSubscription()

	p := mqtt.NewSubscribe()
	// configure settings...

	s.SetPacket(&p)
	s.SetHandler(func(_ context.Context, p proto.Packet) error {
		return fmt.Errorf(": todo")
	})
}
