package mq

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestUndefined(t *testing.T) {
	p := NewUndefined()

	if _, err := p.WriteTo(ioutil.Discard); err == nil {
		t.Error("WriteTo works?!")
	}
	if v := p.String(); !strings.Contains(v, "UNDEFINED") {
		t.Error(v)
	}
}
