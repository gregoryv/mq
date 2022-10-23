package tt

import (
	"context"
	"testing"
	"time"

	"github.com/gregoryv/mq"
)

func TestConnWait(t *testing.T) {
	var (
		conwait     = NewConnWait()
		in          = conwait.In(NoopHandler)
		ctx, cancel = context.WithTimeout(
			context.Background(), 20*time.Millisecond,
		)
	)
	defer cancel()
	{
		p := mq.NewConnAck()
		go in(ctx, &p)
	}
	select {
	case <-conwait.Done():
	case <-ctx.Done():
		t.Error(ctx.Err())
	}
}
