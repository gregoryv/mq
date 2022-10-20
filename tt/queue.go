package tt

import (
	"github.com/gregoryv/mq"
)

func NewQueue(last mq.Handler, v ...mq.Middleware) mq.Handler {
	if len(v) == 0 {
		return last
	}
	l := len(v) - 1
	return v[l](NewQueue(last, v[:l]...))
}
