package tt

import "testing"

func TestNoop(t *testing.T) {
	if err := NoopPub(nil, nil); err != nil {
		t.Fail()
	}
}
