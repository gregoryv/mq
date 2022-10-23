package pink

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	s := NewServer()

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(2*s.acceptTimeout, cancel)

	if err := s.Run(ctx); !errors.Is(err, context.Canceled) {
		t.Error(err)
	}
}
