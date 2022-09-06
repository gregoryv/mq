package client

import (
	"context"
	"testing"

	"github.com/gregoryv/mqtt"
)

func TestAckman(t *testing.T) {
	m := NewAckman(NewIDPool(3))
	ctx := context.Background()
	m.Next(ctx, false) // 1
	m.Next(ctx, true)  // 2

	a := mqtt.NewPubAck()
	a.SetPacketID(2)
	m.Handle(ctx, &a) // Handle packet id 1 should panic
}
