package mqtt

import (
	"bytes"
	"fmt"
	"strings"
)

// FixedHeader represents the first 2..5 bytes of a control packet.

// It's an error if len(FixedHeader) < 2 or > 5.
//
// 2.1.1 Fixed Header
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_MQTT_Control_Packet
type FixedHeader []byte

// String returns a string TYPE-FLAGS REMAINING_LENGTH
func (h FixedHeader) String() string {
	var sb strings.Builder
	sb.WriteString(typeNames[h[0]&0b1111_0000])

	if flags := flagNames.Join("-", h.flagsByValue()); len(flags) > 0 {
		sb.WriteString("-")
		sb.WriteString(flags)
	}
	if rem := h.RemLen(); rem > 1 {
		sb.WriteString(" ")
		fmt.Fprint(&sb, rem)
	}
	return sb.String()
}

// Is is the same as h.Value() == v
func (h FixedHeader) Is(v byte) bool {
	return h.Value() == v
}

func (h FixedHeader) Value() byte {
	return h[0] & 0b1111_0000
}

// RemLen returns the remaining length
func (h FixedHeader) RemLen() int {
	if len(h) < 2 {
		return 0
	}
	v, _ := ParseVarInt(bytes.NewReader(h[1:]))
	return int(v)
}

func (h FixedHeader) HasFlag(f byte) bool {
	return h[0]&f == f
}

func (h FixedHeader) flagsByValue() []byte {
	flags := make([]byte, 0, 4) // max four
	add := func(f ...byte) {
		if len(f) == 1 && h.HasFlag(f[0]) {
			flags = append(flags, f[0])
			return
		}
		if f, ok := h.hasOneFlag(f...); ok {
			flags = append(flags, f)
		}
	}
	builders := map[byte]func(){
		PUBLISH: func() { add(DUP); add(QoS2, QoS1); add(RETAIN) },
	}
	if build, found := builders[h.Value()]; found {
		build()
	}
	return flags
}

func (h FixedHeader) hasOneFlag(flags ...byte) (byte, bool) {
	for _, f := range flags {
		if !h.HasFlag(f) {
			continue
		}
		return f, true
	}
	return 0, false
}
