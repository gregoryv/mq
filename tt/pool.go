package tt

import (
	"context"
)

// Max packet id one client will use starting with 1. This also
// dictates the maximum number of packets in flight.
var MaxDefaultConcurrentID uint16 = 100

// newPool returns a pool of reusable id's from 1..max, 0 is not used
func newPool(max uint16) *pool {
	ids := make(chan uint16, max)
	for i := uint16(1); i <= max; i++ {
		ids <- i
	}

	return &pool{
		pool: ids,
	}
}

type pool struct {
	pool chan uint16
}

// Next returns the next available ID, blocks until one is available
// or context is canceled. Next is safe for concurrent use by multiple
// goroutines.
func (p *pool) Next(ctx context.Context) uint16 {
	select {
	case <-ctx.Done():
		return 0
	default:
		return <-p.pool
	}
}

// Reuse returns the given value to the pool
func (p *pool) Reuse(v uint16) {
	if v == 0 {
		return
	}
	p.pool <- v
}
