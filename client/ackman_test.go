package client

import (
	"context"
	"testing"

	"github.com/gregoryv/mqtt"
)

func TestAckman(t *testing.T) {
	m := NewAckman(NewIDPool(3))
	ctx := context.Background()
	m.Next(ctx)
	m.Next(ctx)

	a := mqtt.NewPubAck()
	a.SetPacketID(1)
	m.Handle(ctx, &a)
}
