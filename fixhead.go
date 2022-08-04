package mqtt

import (
	"fmt"
	"strings"
)

// 2.1.1 Fixed Header
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_MQTT_Control_Packet
type FixedHeader []byte

func (h FixedHeader) String() string {
	str := FlagNames.Join("-", h.FlagsByValue())
	if len(str) > 0 {
		return fmt.Sprintf("%s-%s", h.Name(), str)
	}
	return fmt.Sprintf("%s", h.Name())
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
	return controlPacketTypeName[byte(h[0])&0b1111_0000]
}

func (h FixedHeader) Value() byte {
	return byte(h[0]) & 0b1111_0000
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

func (h FixedHeader) HasFlag(f byte) bool {
	return h.Flags()&f == f
}

func (h FixedHeader) Flags() byte {
	return byte(h[0]) & 0b0000_1111
}

// Fixed header flags
const (
	RETAIN byte = 0b0000_0001
	QoS0   byte = 0b0000_0000
	QoS1   byte = 0b0000_0010
	QoS2   byte = 0b0000_0100
	//QoS3 FixedHeader = 0b0000_0110   malformed!
	DUP byte = 0b0000_1000
)

var FlagNames = ByteNames{
	names: map[byte]string{
		DUP:    "DUP",
		QoS0:   "QoS0",
		QoS1:   "QoS1",
		QoS2:   "QoS2",
		RETAIN: "RETAIN",
	},
}

type ByteNames struct {
	names map[byte]string
}

func (n *ByteNames) Name(b byte) string {
	return n.names[b]
}

func (n *ByteNames) Join(sep string, b []byte) string {
	if len(b) == 0 {
		return ""
	}
	names := make([]string, len(b), len(b))
	for i, b := range b {
		names[i] = n.Name(b)
	}
	return strings.Join(names, sep)
}
