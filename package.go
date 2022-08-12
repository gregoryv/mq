/*
Package mqtt provides a MQTT v5.0 protocol implementation

The specification is found at
https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html

*/
package mqtt

import (
	"encoding"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// ---------------------------------------------------------------------
// Headers
// ---------------------------------------------------------------------

// FixedHeader represents the first 2..5 bytes of a control packet.

// It's an error if len(FixedHeader) < 2 or > 5.
//
// 2.1.1 Fixed Header
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_MQTT_Control_Packet
type FixedHeader struct {
	header       byte
	remainingLen vbint
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
	return bits(f.header).Has(flag)
}

// ---------------------------------------------------------------------
// Data representations, the low level data types
// ---------------------------------------------------------------------

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901013
type spair [2]u8str

func (v spair) WriteTo(w io.Writer) (int64, error) {
	return src(v).WriteTo(w)
}

func (v spair) MarshalBinary() ([]byte, error) {
	key, err := v[0].MarshalBinary()
	if err != nil {
		return nil, marshalErr(v, "key", err.(*Malformed))
	}
	val, err := v[1].MarshalBinary()
	if err != nil {
		return nil, marshalErr(v, "value", err.(*Malformed))
	}
	return append(key, val...), nil
}

func (v *spair) UnmarshalBinary(data []byte) error {
	if err := v[0].UnmarshalBinary(data); err != nil {
		return unmarshalErr(v, "key", err.(*Malformed))
	}
	i := len(v[0]) + 2
	if err := v[1].UnmarshalBinary(data[i:]); err != nil {
		return unmarshalErr(v, "value", err.(*Malformed))
	}
	return nil
}
func (v spair) String() string {
	return fmt.Sprintf("%s:%s", v[0], v[1])
}

// ----------------------------------------

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901010
type u8str string

func (v u8str) WriteTo(w io.Writer) (int64, error) {
	return src(v).WriteTo(w)
}

func (v u8str) MarshalBinary() ([]byte, error) {
	data, err := bindat([]byte(v)).MarshalBinary()
	if err != nil {
		return data, marshalErr(v, "", err.(*Malformed))
	}
	return data, nil
}

func (v *u8str) UnmarshalBinary(data []byte) error {
	var b bindat
	if err := b.UnmarshalBinary(data); err != nil {
		return unmarshalErr(v, "", err.(*Malformed))
	}
	*v = u8str(b)
	return nil
}

// ----------------------------------------

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901012
type bindat []byte

func (v bindat) WriteTo(w io.Writer) (int64, error) {
	return src(v).WriteTo(w)
}

func (v bindat) MarshalBinary() ([]byte, error) {
	if len(v) > MaxUint16 {
		return nil, marshalErr(v, "", "size exceeded")
	}
	data := make([]byte, len(v)+2)
	l, _ := b2int(len(v)).MarshalBinary()
	copy(data[:2], l)
	copy(data[2:], []byte(v))
	return data, nil
}

func (v *bindat) UnmarshalBinary(data []byte) error {
	var l b2int
	_ = l.UnmarshalBinary(data)
	if len(data) < int(l)+2 {
		return unmarshalErr(v, "", "missing data")
	}
	*v = make([]byte, l)
	copy(*v, data[2:l+2])
	return nil
}

// ----------------------------------------

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901011
type vbint uint

func (v vbint) WriteTo(w io.Writer) (int64, error) {
	return src(v).WriteTo(w)
}

// MarshalBinary always returns nil error
func (v vbint) MarshalBinary() ([]byte, error) {
	data := make([]byte, 0, 4) // max four
	if v == 0 {
		data = append(data, 0)
		return data, nil
	}
	for v > 0 {
		encodedByte := byte(v % 128)
		v = v / 128
		if v > 0 {
			encodedByte = encodedByte | 128
		}
		data = append(data, encodedByte)
	}
	return data, nil
}

// UnmarshalBinary data, returns nil or *Malformed error
func (v *vbint) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return unmarshalErr(v, "", "missing data")
	}
	var multiplier uint = 1
	var value uint
	for _, encodedByte := range data {
		value += uint(encodedByte) & uint(127) * multiplier
		if multiplier > 128*128*128 {
			return unmarshalErr(v, "", "size exceeded")
		}
		if encodedByte&128 == 0 {
			break
		}
		multiplier = multiplier * 128
	}
	*v = vbint(value)
	return nil
}

func (v vbint) Width() int {
	switch {
	case v < 128:
		return 1
	case v < 16_384:
		return 2
	case v < 2_097_152:
		return 3
	default:
		return 4
	}
}

// ----------------------------------------

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901007
type bits byte

func (v bits) Has(b byte) bool { return byte(v)&b == b }

// ----------------------------------------

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901008
type b2int uint16

func (v b2int) WriteTo(w io.Writer) (int64, error) {
	return src(v).WriteTo(w)
}

func (v b2int) MarshalBinary() ([]byte, error) {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, uint16(v))
	return data, nil
}

func (v *b2int) UnmarshalBinary(data []byte) error {
	*v = b2int(binary.BigEndian.Uint16(data))
	return nil
}

// ----------------------------------------

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901009
type b4int uint32

func (v b4int) WriteTo(w io.Writer) (int64, error) {
	return src(v).WriteTo(w)
}

func (v b4int) MarshalBinary() ([]byte, error) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, uint32(v))
	return data, nil
}

func (v *b4int) UnmarshalBinary(data []byte) error {
	*v = b4int(binary.BigEndian.Uint32(data))
	return nil
}

// ----------------------------------------

func src(v encoding.BinaryMarshaler) io.WriterTo {
	return writerToFunc(func(w io.Writer) (int64, error) {
		data, err := v.MarshalBinary()
		if err != nil {
			return 0, err
		}
		n, err := w.Write(data)
		return int64(n), err
	})
}

type writerToFunc func(w io.Writer) (int64, error)

func (f writerToFunc) WriteTo(w io.Writer) (int64, error) {
	return f(w)
}

// ---------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------

func marshalErr(v interface{}, ref string, err interface{}) *Malformed {
	e := newMalformed(v, ref, err)
	e.method = "marshal"
	return e
}

func unmarshalErr(v interface{}, ref string, err interface{}) *Malformed {
	e := newMalformed(v, ref, err)
	e.method = "unmarshal"
	return e
}

func newMalformed(v interface{}, ref string, err interface{}) *Malformed {
	var reason string
	switch e := err.(type) {
	case *Malformed:
		reason = e.reason
	case string:
		reason = e
	}
	// remove * from type name
	t := fmt.Sprintf("%T", v)
	if t[0] == '*' {
		t = t[1:]
	}
	return &Malformed{
		t:      t,
		ref:    ref,
		reason: reason,
	}
}

type Malformed struct {
	method string
	t      string
	ref    string
	reason string
}

func (e *Malformed) Error() string {
	if e.ref == "" {
		return fmt.Sprintf("malformed %s %s: %s", e.t, e.method, e.reason)
	}
	return fmt.Sprintf("malformed %s %s: %s %s", e.t, e.method, e.ref, e.reason)
}

// ---------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------

const (
	MQTT      = "MQTT" // 3.1.2.1 Protocol Name
	Version5  = 5
	MaxUint16 = 1<<16 - 1
)

// 2.1.2 MQTT Control Packet type
//
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_MQTT_Control_Packet
const (
	UNDEFINED   byte = (iota << 4) // 0 Forbidden Reserved
	CONNECT                        // 1 Client to Server Connection request
	CONNACK                        // 2 Server to Client Connect acknowledgment
	PUBLISH                        // 3 Client to Server or Publish message
	PUBACK                         // 4 Client to Server or Publish acknowledgment (QoS 1)
	PUBREC                         // 5 Client to Server or Publish received (QoS 2 delivery part 1)
	PUBREL                         // 6 Client to Server or Publish release (QoS 2 delivery part 2)
	PUBCOMP                        // 7 Client to Server or Publish complete (QoS 2 delivery part 3)
	SUBSCRIBE                      // 8 Client to Server Subscribe request
	SUBACK                         // 9 Server to Client Subscribe acknowledgment
	UNSUBSCRIBE                    // 10 Client to Server Unsubscribe request
	UNSUBACK                       // 11 Server to Client Unsubscribe acknowledgment
	PINGREQ                        // 12 Client to Server PING request
	PINGRESP                       // 13 Server to Client PING response
	DISCONNECT                     // 14 Client to Server or Disconnect notification
	AUTH                           // 15 Client to Server or Server to Client Authentication exchange
)

var typeNames = map[byte]string{
	UNDEFINED:   "UNDEFINED",
	CONNECT:     "CONNECT",
	CONNACK:     "CONNACK",
	PUBLISH:     "PUBLISH",
	PUBACK:      "PUBACK",
	PUBREC:      "PUBREC",
	PUBREL:      "PUBREL",
	PUBCOMP:     "PUBCOMP",
	SUBSCRIBE:   "SUBSCRIBE",
	SUBACK:      "SUBACK",
	UNSUBSCRIBE: "UNSUBSCRIBE",
	UNSUBACK:    "UNSUBACK",
	PINGREQ:     "PINGREQ",
	PINGRESP:    "PINGRESP",
	DISCONNECT:  "DISCONNECT",
	AUTH:        "AUTH",
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

const (
	PropSessionExpiryInterval byte = 0x11
	PropReceiveMax            byte = 0x21
	PropMaxPacketSize         byte = 0x27
)

// The Reason Codes used for Malformed Packet and Protocol Errors
//
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Handling_errors
const (
	MalformedPacket                     byte = 0x81 // Malformed Packet
	ProtocolError                       byte = 0x82 // Protocol Error
	ReceiveMaximumExceeded              byte = 0x93 // Receive Maximum exceeded
	PacketTooLarge                      byte = 0x95 // Packet too large
	RetainNotSupported                  byte = 0x9A // Retain not supported
	QoSNotSupported                     byte = 0x9B // QoS not supported
	SharedSubscriptionsNotSupported     byte = 0x9E // Shared Subscriptions not supported
	SubscriptionIdentifiersNotSupported byte = 0xA1 // Subscription Identifiers not supported
	WildcardSubscriptionsNotSupported   byte = 0xA2 // Wildcard Subscriptions not supported
)

var codeNames = map[byte]string{
	MalformedPacket:                     "Malformed Packet",
	ProtocolError:                       "Protocol Error",
	ReceiveMaximumExceeded:              "Receive Maximum exceeded",
	PacketTooLarge:                      "Packet too large",
	RetainNotSupported:                  "Retain not supported",
	QoSNotSupported:                     "QoS not supported",
	SharedSubscriptionsNotSupported:     "Shared Subscriptions not supported",
	SubscriptionIdentifiersNotSupported: "Subscription Identifiers not supported",
	WildcardSubscriptionsNotSupported:   "Wildcard Subscriptions not supported",
}
