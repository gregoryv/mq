package tt

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/gregoryv/mq"
)

func TestReceiver(t *testing.T) {
	conn, server := Dial()

	var wg sync.WaitGroup
	wg.Add(1)
	receiver := NewReceiver(conn, func(_ context.Context, _ mq.Packet) error {
		wg.Done()
		return nil
	})

	ctx := context.Background()
	go receiver.Run(ctx)
	p := mq.NewPublish()
	p.WriteTo(server)

	wg.Wait()
}

func TestReceiver_RunRespectsContextCancel(t *testing.T) {
	conn := dialBroker(t)
	receiver := NewReceiver(conn, NoopHandler)
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

func TestReceiver_closedConn(t *testing.T) {
	receiver := NewReceiver(&ClosedConn{}, NoopHandler)

	var wg sync.WaitGroup
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Millisecond)

	wg.Add(1)
	go func() {
		if err := receiver.Run(ctx); !errors.Is(err, io.EOF) {
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
