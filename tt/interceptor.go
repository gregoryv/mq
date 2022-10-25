package tt

import (
	"context"

	"github.com/gregoryv/mq"
)

// Intercept returns an interceptor of any packet.
func Intercept[T mq.Packet](c chan T) InFlow {
	return InFunc(func(next mq.Handler) mq.Handler {
		return func(ctx context.Context, p mq.Packet) error {
			switch p := p.(type) {
			case T:
				c <- p
			}
			return nil
		}
	})
}
