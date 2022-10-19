package tt

import (
	"github.com/gregoryv/mq"
)

func NewQueue(v []mq.Middleware, last mq.Handler) mq.Handler {
	if len(v) == 0 {
		return last
	}
	return v[0](NewQueue(v[1:], last))
}
