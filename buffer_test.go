package mqtt

import "testing"

func Test_buffer(t *testing.T) {
	b := &buffer{}
	b.getAny(map[Ident]wireType{}, func(property) {})
	if b.err == nil {
		t.Error("expect getAny to fail on missing data")
	}

	b = &buffer{data: []byte{2, 0xff, 0}}
	b.getAny(map[Ident]wireType{}, func(property) {})
	if b.err == nil {
		t.Error("expect getAny to fail")
	}

	b = &buffer{data: []byte{2}}
	var aboolean wbool
	if b.get(&aboolean); b.err == nil {
		t.Error("expect get to fail")
	}
}
