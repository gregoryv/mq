package mqtt

import "testing"

var _ wireType = &Publish{}

func Test(t *testing.T) {
	_ = NewPublish()
}
