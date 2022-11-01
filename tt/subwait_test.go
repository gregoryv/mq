package tt

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gregoryv/mq"
)

func TestSubWait(t *testing.T) {
	subscriptions := 3
	var (
		subwait     = NewSubWait(subscriptions)
		in          = subwait.In(NoopHandler)
		out         = subwait.Out(NoopHandler)
		ctx, cancel = context.WithTimeout(
			context.Background(), 20*time.Millisecond,
		)
	)
	defer cancel()

	{
		p := mq.NewConnect()
		_ = out(ctx, p)
	}
	{
		p := mq.NewSubAck()
		for i := 0; i < subscriptions; i++ {
			_ = in(ctx, p)
		}
	}
	select {
	case <-subwait.Done(ctx):
	case <-ctx.Done():
		t.Error(ctx.Err())
	}

	// check timeout
	{
		_ = in(ctx, mq.NewSubAck()) // only send one
	}
	select {
	case <-subwait.Done(ctx):
	case <-ctx.Done():
		if err := ctx.Err(); !errors.Is(err, context.DeadlineExceeded) {
			t.Error(err)
		}
	}

}
