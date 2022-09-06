package client

import (
	"context"
	"sync"
)

// NewIDPool returns a pool of reusable id's from 1..max, 0 is not
// used
func NewIDPool(max uint16) *IDPool {
	return &IDPool{
		pool:     make([]bool, max),
		lastFree: make(chan uint16),
	}
}

type IDPool struct {
	i    int
	m    sync.Mutex
	pool []bool

	// last id that can be reused
	lastFree chan uint16
}

func (p *IDPool) Next(ctx context.Context) uint16 {
	for {
		select {
		case <-ctx.Done():
			return 0
		default:
		}
		// next in line is most likely free
		width := len(p.pool)
		for i := p.i; i < len(p.pool); i++ {
			if p.pool[i] == FREE {
				p.m.Lock()
				p.pool[i] = USED
				p.m.Unlock()
				p.i = i + 1 // ready for next
				return uint16(i + 1)
			}
			width--
		}
		p.i = 0
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

func (p *IDPool) Reuse(v uint16) {
	p.m.Lock()
	p.pool[v-1] = FREE
	p.m.Unlock()
	select {
	case p.lastFree <- v:
	default:
		// nobody is waiting for it
	}
}

var USED = true
var FREE = false
