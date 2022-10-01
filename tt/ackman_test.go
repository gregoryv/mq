package tt

import (
	"context"
	"testing"
)

func Test_ackman(t *testing.T) {
	// using a pool of maximum 3 packet ids, 1,2 and 3
	m := newAckman(NewIDPool(3))
	ctx := context.Background()
	m.Next(ctx)         // 1
	last := m.Next(ctx) // 2

	if err := m.Handle(ctx, last); err != nil {
		t.Error(err)
	}

	notUsed := uint16(3)
	if err := m.Handle(ctx, notUsed); err == nil {
		t.Error("expect error when trying to handle free packet id")
	}
}
