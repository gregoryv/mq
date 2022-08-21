package mqtt

import (
	"fmt"
)

type buffer struct {
	data []byte
	i    int
	err  error
}

func (b *buffer) getAny(fields map[Ident]wireType, addProp func(property)) {
	var propLen vbint
	b.get(&propLen)
	end := b.i + int(propLen)
	var id Ident
	for b.i < end {
		before := b.i
		b.get(&id)
		if b.i == before {
			return
		}
		field, hasField := fields[id]
		switch {
		case hasField:
			b.get(field)

		case id == UserProperty:
			var p property
			b.get(&p)
			addProp(p)

		default:
			b.err = fmt.Errorf("unknown property id 0x%02x", id)
		}
	}
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
