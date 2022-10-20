package tt

import (
	"context"
	"io"
	"sync"

	"github.com/gregoryv/mq"
)

func NewSender(v io.Writer) *Sender {
	return &Sender{Writer: v}
}

type Sender struct {
	sync.Mutex
	io.Writer
}

// Out writes the packet to the underlying writer. Safe for
// concurrent calls.
func (s *Sender) Out(ctx context.Context, p mq.Packet) error {
	s.Lock()
	_, err := p.WriteTo(s.Writer)
	s.Unlock()
	return err
}
