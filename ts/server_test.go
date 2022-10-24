package ts

import (
	. "context"
	"errors"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	s := NewServer()

	ctx, cancel := WithCancel(Background())
	time.AfterFunc(2*s.acceptTimeout, cancel)

	if err := s.Run(ctx); !errors.Is(err, Canceled) {
		t.Error(err)
	}
}
