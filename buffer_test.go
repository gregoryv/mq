package mq

import (
	"errors"
	"testing"
)

func Test_buffer(t *testing.T) {
	{ // missing data
		b := &buffer{}
		b.getAny(map[Ident]func() wireType{}, func(UserProp) {})
		if b.err != nil {
			t.Error("getAny failes on empty data")
		}
	}
	{ // unknown user property
		b := &buffer{data: []byte{2, 0xff, 0}}
		b.getAny(map[Ident]func() wireType{}, func(UserProp) {})
		if b.err == nil {
			t.Error("expect getAny to fail")
		}
	}
	{ // invalid data type, 0 or 1 are accepted
		b := &buffer{data: []byte{2}}
		var aboolean wbool
		if b.get(&aboolean); b.err == nil {
			t.Error("expect get to fail")
		}
	}
	{ // missing data
		b := &buffer{data: []byte{}}
		var binary bindata
		if b.get(&binary); !errors.Is(b.err, ErrMissingData) {
			t.Error("expect get to fail")
		}
	}
}
