package mqtt

import (
	"fmt"
	"strings"
)

// 2.1.1 Fixed Header
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_MQTT_Control_Packet
type FixedHeader []byte

func (h FixedHeader) String() string {
	// todo flags by packet type
	f := make([]byte, 0, 4) // max four
	if h.HasFlag(DUP) {
		f = append(f, DUP)
	}
	if h.HasFlag(RETAIN) {
		f = append(f, RETAIN)
	}

	str := flags.Join("-", f)
	if len(str) > 0 {
		return fmt.Sprintf("%s-%s", h.Name(), str)
	}
	return fmt.Sprintf("%s", h.Name())
}

func (h FixedHeader) Name() string {
	return controlPacketTypeName[byte(h[0])&0b1111_0000]
}

func (h FixedHeader) Value() byte {
	return byte(h[0]) & 0b1111_0000
}

func (h FixedHeader) HasFlag(f byte) bool {
	return h.Flags()&f == f
}

func (h FixedHeader) Flags() byte {
	return byte(h[0]) & 0b0000_1111
}

// FixedHeader flags
const (
	DUP    byte = 0b0000_1000
	RETAIN byte = 0b0000_0001

	QoS0 byte = 0b0000_0000
	QoS1 byte = 0b0000_0010
	QoS2 byte = 0b0000_0100
	//QoS3 FixedHeader = 0b0000_0110   malformed!
)

var flags = ByteNames{
	names: map[byte]string{
		DUP:    "DUP",
		RETAIN: "RETAIN",

		QoS0: "QoS0",
		QoS1: "QoS1",
		QoS2: "QoS2",
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
