package pakio

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/gregoryv/mq"
)

func TestReceiver_RunRespectsContextCancel(t *testing.T) {
	conn := dialBroker(t)
	receiver := NewReceiver(conn, mq.NoopHandler)
	var wg sync.WaitGroup
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Millisecond)

	wg.Add(1)
	go func() {
		if err := receiver.Run(ctx); !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("unexpected error: %T", err)
		}
		wg.Done()
	}()

	wg.Wait()
}

func dialBroker(t *testing.T) net.Conn {
	conn, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		t.Skip(err)
		return nil
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}
