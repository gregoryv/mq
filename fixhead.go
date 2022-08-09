package mqtt

import (
	"fmt"
	"strings"
)

// FixedHeader represents the first 2..5 bytes of a control packet.

// It's an error if len(FixedHeader) < 2 or > 5.
//
// 2.1.1 Fixed Header
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_MQTT_Control_Packet
type FixedHeader struct {
	header       byte
	remainingLen VarByteInt
}

func (f *FixedHeader) MarshalBinary() ([]byte, error) {
	data := make([]byte, 0, 5)
	data = append(data, f.header)
	rem, _ := f.remainingLen.MarshalBinary() // cannot fail
	data = append(data, rem...)
	return data, nil
}

func (f *FixedHeader) UnmarshalBinary(data []byte) error {
	f.header = data[0]
	err := f.remainingLen.UnmarshalBinary(data[1:])
	if err != nil {
		return unmarshalErr(f, "remaining length", err.(*Malformed))
	}
	return nil
}

// String returns a string TYPE-FLAGS REMAINING_LENGTH
func (f *FixedHeader) String() string {
	var sb strings.Builder
	sb.WriteString(typeNames[f.Value()])
	sb.WriteString(" ")
	flags := []byte("----")
	if f.HasFlag(DUP) {
		flags[0] = 'd'
	}
	switch {
	case f.HasFlag(QoS1 | QoS2):
		flags[1] = '!' // malformed
		flags[2] = '!' // malformed
	case f.HasFlag(QoS1):
		flags[2] = '1'
	case f.HasFlag(QoS2):
		flags[1] = '2'
	}
	if f.HasFlag(RETAIN) {
		flags[3] = 'r'
	}
	sb.Write(flags)
	sb.WriteString(" ")
	fmt.Fprint(&sb, f.remainingLen)
	return sb.String()
}

// Is is the same as h.Value() == v
func (f *FixedHeader) Is(v byte) bool {
	return f.Value() == v
}

func (f *FixedHeader) Value() byte {
	return f.header & 0b1111_0000
}

func (f *FixedHeader) HasFlag(flag byte) bool {
	return Bits(f.header).Has(flag)
}
