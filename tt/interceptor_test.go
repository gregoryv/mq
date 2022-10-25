package tt

import (
	. "context"
	"testing"
	"time"

	"github.com/gregoryv/mq"
)

func TestIntercept(t *testing.T) {
	i := Intercept[*mq.Connect]()
	h := i.In(NoopHandler)
	go h(Background(), &mq.Connect{})

	select {
	case <-i.C:
	case <-time.After(1 * time.Millisecond):
		t.Fail()
	}
}
