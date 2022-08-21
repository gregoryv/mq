package mqtt

import (
	"bytes"
	"testing"
	"unsafe"
)

var _ wireType = &Publish{}

func TestPublish(t *testing.T) {
	p := NewPublish()
	t.Log(p, unsafe.Sizeof(p))

	var buf bytes.Buffer
	if _, err := p.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	if got := p.width(); got < 0 {
		t.Error(got)
	}
}
