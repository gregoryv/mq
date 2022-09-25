package x

import (
	"context"
	"testing"
)

func TestIDPool(t *testing.T) {
	p := NewIDPool(3) // 1 .. 3

	ctx := context.Background()
	p.Next(ctx) // 1
	p.Next(ctx) // 2
	p.Reuse(2)
	if v := p.Next(ctx); v != 3 {
		t.Error("unexpected next packet id", v)
	}

	// check waiting for
	p.Next(ctx) // 3
	go p.Reuse(3)
	if v := p.Next(ctx); v != 3 {
		t.Error(v)
	}

	p.Reuse(1)
	if v := p.Next(ctx); v != 1 {
		t.Error("unexpected packet id", v)
	}
}
