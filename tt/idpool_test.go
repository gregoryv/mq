package tt

import (
	"context"
	"testing"
	"time"

	"github.com/gregoryv/mq"
)

func Test_IDPool(t *testing.T) {
	max := uint16(5)
	pool := NewIDPool(max) // 1 .. 5

	// check that ids are reusable
	used := make(chan uint16, max)
	drain := func() {
		for v := range used {
			pool.reuse(v)
		}
	}
	ctx := context.Background()
	for i := uint16(0); i < 2*max; i++ {
		v := pool.next(ctx)
		used <- v
		if i == max-1 {
			// start returning midway
			go drain()
		}
	}
	go drain()

	// check all packets that require id
	packets := []mq.Packet{
		mq.Pub(1, "a/b", "gopher"),
		mq.NewSubscribe(),
		mq.NewUnsubscribe(),
		func() mq.Packet {
			p := mq.NewPubAck()
			p.SetPacketID(1)
			return p
		}(),
		func() mq.Packet {
			p := mq.NewPubComp()
			p.SetPacketID(1)
			return p
		}(),
	}

	for _, p := range packets {
		if err := pool.Out(NoopHandler)(ctx, p); err != nil {
			t.Error(err)
		}
		if p, ok := p.(mq.HasPacketID); ok {
			if p.PacketID() == 0 {
				t.Error(p)
			}
		}
		if err := pool.In(NoopHandler)(ctx, p); err != nil {
			t.Error(err)
		}
	}

	// not return 0 value
	pool.reuse(0) // noop
}

func TestIDPool_nextTimeout(t *testing.T) {
	pool := NewIDPool(1) // 1 .. 5
	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond)
	pool.next(ctx)
	if v := pool.next(ctx); v != 0 {
		t.Error("expect 0 id when pool is cancelled", v)
	}
}
