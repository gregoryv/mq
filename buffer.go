package mqtt

import (
	"fmt"
)

// buffer is a much simpler bytes.Buffer which guards sequential
// access to data on error.
type buffer struct {
	data []byte
	i    int // current offset
	err  error

	addSubscriptionID func(uint32) // used in e.g. Publish
}

// getAny reads all properties from the current offset starting with
// the variable length.  fields map property identity codes to wire
// type fields and the addProp func is used for each user property.
func (b *buffer) getAny(fields map[Ident]wireType, addProp func(property)) {
	var propLen vbint
	b.get(&propLen)
	if b.err != nil {
		return
	}
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

		case id == SubscriptionID:
			var sub vbint
			b.get(&sub)
			if b.addSubscriptionID != nil {
				b.addSubscriptionID(uint32(sub))
			}

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
