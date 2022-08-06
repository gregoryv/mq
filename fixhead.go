package mqtt

import (
	"bytes"
	"fmt"
	"strings"
)

// 2.1.1 Fixed Header
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_MQTT_Control_Packet
type FixedHeader []byte

// String returns a string TYPE-FLAGS REMAINING_LENGTH
func (h FixedHeader) String() string {
	var sb strings.Builder
	sb.WriteString(h.Name())

	if flags := flagNames.Join("-", h.FlagsByValue()); len(flags) > 0 {
		sb.WriteString("-")
		sb.WriteString(flags)
	}
	if rem := h.RemLen(); rem > 1 {
		sb.WriteString(" ")
		fmt.Fprint(&sb, rem)
	}
	return sb.String()
}

func (h FixedHeader) FlagsByValue() []byte {
	flags := make([]byte, 0, 4) // max four
	add := func(f ...byte) {
		if len(f) == 1 && h.HasFlag(f[0]) {
			flags = append(flags, f[0])
			return
		}
		if f, ok := h.HasOneFlag(f...); ok {
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

func (h FixedHeader) Name() string {
	return controlPacketTypeName[byte(h.byte1())&0b1111_0000]
}

// Is is the same as h.Value() == v
func (h FixedHeader) Is(v byte) bool {
	return h.Value() == v
}

func (h FixedHeader) Value() byte {
	return byte(h.byte1()) & 0b1111_0000
}

func (h FixedHeader) HasOneFlag(flags ...byte) (byte, bool) {
	for _, f := range flags {
		if !h.HasFlag(f) {
			continue
		}
		return f, true
	}
	return 0, false
}

// RemLen returns the remaining length
func (h FixedHeader) RemLen() uint {
	if len(h) < 2 {
		return 0
	}
	v, _ := ParseVarInt(bytes.NewReader(h[1:]))
	return v
}

func (h FixedHeader) HasFlag(f byte) bool {
	return h.Flags()&f == f
}

func (h FixedHeader) Flags() byte {
	return byte(h.byte1()) & 0b0000_1111
}

func (h FixedHeader) byte1() byte {
	if len(h) == 0 {
		return 0
	}
	return h[0]
}
