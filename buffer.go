package mqtt

import (
	"fmt"
)

type fields map[Ident]wireType

func (f fields) getAny(buf *buffer, addProp func(property)) {
	var propLen vbint
	get := buf.get
	get(&propLen)
	end := buf.Index() + int(propLen)
	var id Ident
	for buf.Index() < end {
		get(&id)
		field, hasField := f[id]
		switch {
		case hasField:
			get(field)

		case id == UserProperty:
			var p property
			get(&p)
			addProp(p)

		default:
			buf.err = fmt.Errorf("unknown property id 0x%02x", id)
		}
	}
}

type buffer struct {
	data []byte
	i    int
	err  error
}

func (b *buffer) get(v wireType) {
	if b.err != nil {
		return
	}
	if b.err = v.UnmarshalBinary(b.data[b.i:]); b.err != nil {
		return
	}
	b.i += v.width()
}

func (b *buffer) Err() error { return b.err }
func (b *buffer) Index() int { return b.i }
