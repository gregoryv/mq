package tt

import (
	"context"
	"sync"
)

// newPool returns a pool of reusable id's from 1..max, 0 is not used
func newPool(max uint16) *pool {
	return &pool{
		pool:     make([]bool, max),
		lastFree: make(chan uint16),
	}
}

type pool struct {
	nextFreeIndex int
	m             sync.RWMutex
	pool          []bool

	// last id that can be reused
	lastFree chan uint16
}

func (p *pool) Next(ctx context.Context) uint16 {
	for {
		select {
		case <-ctx.Done():
			return 0
		default:
		}
		width := len(p.pool)

		// next in line is most likely free
		for i := p.nextFreeIndex; i < len(p.pool); i++ {
			if p.pool[i] == FREE {
				p.m.Lock()
				p.pool[i] = USED
				p.m.Unlock()
				p.nextFreeIndex = i + 1 // ready for next
				return uint16(i + 1)
			}
			width--
		}
		p.nextFreeIndex = 0
		if width == 0 {
			// all ids are being used, wait for next free
			select {
			case <-ctx.Done():
			case v := <-p.lastFree:
				return v
			}
		}
	}
}

// InUse returns true if the given value is not in the pool at this
// moment.
func (p *pool) InUse(v uint16) bool {
	p.m.RLock()
	u := p.pool[v-1]
	p.m.RUnlock()
	return u
}

// Reuse returns the given value to the pool
func (p *pool) Reuse(v uint16) {
	p.m.Lock()
	p.pool[v-1] = FREE
	p.m.Unlock()
	select {
	case p.lastFree <- v:
	default:
		// nobody is waiting for it
	}
}

var (
	USED = true
	FREE = false
)
