package tt

import (
	"context"
	"testing"

	"github.com/gregoryv/mq"
)

func TestAckman(t *testing.T) {
	// using a pool of maximum 3 packet ids, 1,2 and 3
	m := NewAckman(NewIDPool(3))
	ctx := context.Background()
	m.Next(ctx)         // 1
	last := m.Next(ctx) // 2

	a := mq.NewPubAck()
	a.SetPacketID(last)
	if err := m.Handle(ctx, &a); err != nil {
		t.Error(err)
	}

	a.SetPacketID(3) // not used
	if err := m.Handle(ctx, &a); err == nil {
		t.Error("expect error when trying to handle free packet id")
	}
}
