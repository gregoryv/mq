package tt

import (
	"context"
	"sync"
	"testing"

	"github.com/gregoryv/mq"
)

func TestSubscriber(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	s := NewSubscriber(func(_ context.Context, p mq.Packet) error {
		_ = p.(*mq.Subscribe)
		wg.Done()
		return nil
	},
		NewRoute("#", NoopPub),
	)

	p := mq.NewConnAck()
	go s.SubscribeOnConnect(NoopHandler)(nil, &p)

	wg.Wait()
}
