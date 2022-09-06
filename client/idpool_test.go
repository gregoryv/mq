package client

import (
	"context"
	"testing"
)

func TestIDPool(t *testing.T) {
	p := NewIDPool(3) // 1 .. 3

	ctx := context.Background()
	p.Next(ctx) // 1
	p.Next(ctx) // 2
	p.Next(ctx) // 3
	p.Reuse(2)
	if v := p.Next(ctx); v != 2 {
		t.Error(v)
	}

	// check waiting for
	go p.Reuse(2)
	if v := p.Next(ctx); v != 2 {
		t.Error(v)
	}
}
