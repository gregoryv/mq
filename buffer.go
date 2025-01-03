package mq

import (
	"fmt"
)

// buffer is a much simpler bytes.Buffer which guards sequential
// access to data on error.
//
// Note! The buffer is only used for reading as filling the buffer
// increased memory allocation, due to the fact we need partial widths
// of e.g. properties or payloads. If used in the fill methods, that
// would mean we'd have to allocate multiple of them.
type buffer struct {
	data []byte
	i    int // current offset
	err  error

	addSubscriptionID func(uint32) // used in e.g. Publish
}

// getAny reads all properties from the current offset starting with
// the variable length.  fields map property identity codes to wire
// type fields and the addProp func is used for each user property.
func (b *buffer) getAny(fields map[Ident]func() wireType, addProp func(UserProp)) {
	if b.atEnd() {
		return
	}
	var propLen vbint
	b.get(&propLen)
	end := b.i + int(propLen)
	var id Ident
	for b.i < end {
		b.get(&id)
		// first failure stops the parsing
		if b.err != nil {
			return
		}
		field, hasField := fields[id]
		if hasField {
			b.get(field())
			continue
		}
		switch id {
		case UserProperty:
			var p UserProp
			b.get(&p)
			addProp(p)

		case SubscriptionID:
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
	if b.i >= len(b.data) {
		b.err = ErrMissingData
		return
	}
	if b.err = v.UnmarshalBinary(b.data[b.i:]); b.err != nil {
		return
	}
	b.i += v.width()
}

func (b *buffer) atEnd() bool {
	return b.i == len(b.data)
}

func (b *buffer) Err() error { return b.err }

var ErrMissingData = fmt.Errorf("missing data")
