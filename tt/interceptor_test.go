package tt

import (
	. "context"
	"testing"
	"time"

	"github.com/gregoryv/mq"
)

func TestIntercept(t *testing.T) {
	c := make(chan *mq.Connect, 0)
	i := Intercept(c)
	h := i.In(NoopHandler)
	go h(Background(), &mq.Connect{})

	select {
	case <-c:
	case <-time.After(1 * time.Millisecond):
		t.Fail()
	}
}
