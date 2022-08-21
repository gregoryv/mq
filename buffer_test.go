package mqtt

import "testing"

func Test_buffer(t *testing.T) {
	var b buffer

	if b.err != nil {
		t.Fatal("empty buffer cannot have error")
	}

	b.data = []byte{2, 0xff, 0}
	b.getAny(map[Ident]wireType{}, func(property) {})

	b.i = 0
	b.data = []byte{2}
	b.err = nil
	var aboolean wbool
	if b.get(&aboolean); b.err == nil {
		t.Error("get wbool with bad data should fail")
	}
}
