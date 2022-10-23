package pink

import (
	"context"
	"errors"
	"testing"
)

func TestServer(t *testing.T) {
	s := NewServer()

	ctx, cancel := context.WithCancel(context.Background())
	go cancel()

	if err := s.Run(ctx); !errors.Is(err, context.Canceled) {
		t.Error(err)
	}
}
