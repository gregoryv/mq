package tt

import (
	"github.com/gregoryv/mq"
)

func NewInQueue(last mq.Handler, v ...InFlow) mq.Handler {
	if len(v) == 0 {
		return last
	}
	l := len(v) - 1
	return v[l].In(NewInQueue(last, v[:l]...))
}

func NewOutQueue(last mq.Handler, v ...OutFlow) mq.Handler {
	if len(v) == 0 {
		return last
	}
	l := len(v) - 1
	return v[l].Out(NewOutQueue(last, v[:l]...))
}
