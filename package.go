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

// Check verifies if the given packet is well formed, returns
// Malformed error if not.
func Check(p Packet) error {
	return p.check()
}

type Packet interface {
	check() error
}

// ---------------------------------------------------------------------
// Headers
// ---------------------------------------------------------------------

// firstByte represents the first byte in a control packet.
type firstByte byte

// String returns a string TYPE-FLAGS REMAINING_LENGTH
func (f firstByte) String() string {
	var sb strings.Builder
	sb.WriteString(typeNames[byte(f)&0b1111_0000])
	sb.WriteString(" ")
	flags := []byte("----")
	if bits(f).Has(DUP) {
		flags[0] = 'd'
	}
	switch {
	case bits(f).Has(QoS1 | QoS2):
		flags[1] = '!' // malformed
		flags[2] = '!' // malformed
	case bits(f).Has(QoS1):
		flags[2] = '1'
	case bits(f).Has(QoS2):
		flags[1] = '2'
	}
	if bits(f).Has(RETAIN) {
		flags[3] = 'r'
	}
	sb.Write(flags)
	return sb.String()
}

// ---------------------------------------------------------------------
// Data representations, the low level data types
// ---------------------------------------------------------------------

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901013
type property [2]u8str

func (v property) WriteTo(w io.Writer) (int64, error) {
	return src(v).WriteTo(w)
}

func (v property) MarshalBinary() ([]byte, error) {
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

func (v property) MarshalInto(data []byte) {
	_ = data[v.width()-1]
	v[0].MarshalInto(data)
	i := v[0].width()
	v[1].MarshalInto(data[i:])
}

func (v *property) UnmarshalBinary(data []byte) error {
	if err := v[0].UnmarshalBinary(data); err != nil {
		return unmarshalErr(v, "key", err.(*Malformed))
	}
	i := len(v[0]) + 2
	if err := v[1].UnmarshalBinary(data[i:]); err != nil {
		return unmarshalErr(v, "value", err.(*Malformed))
	}
	return nil
}
func (v property) String() string {
	return fmt.Sprintf("%s:%s", v[0], v[1])
}
func (v property) width() int {
	return v[0].width() + v[1].width()
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

func (v u8str) MarshalInto(data []byte) {
	bindat(v).MarshalInto(data)
}

func (v *u8str) UnmarshalBinary(data []byte) error {
	var b bindat
	if err := b.UnmarshalBinary(data); err != nil {
		return unmarshalErr(v, "", err.(*Malformed))
	}
	*v = u8str(b)
	return nil
}

func (v u8str) width() int {
	return bindat(v).width()
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
	data := make([]byte, v.width())
	v.MarshalInto(data)
	return data, nil
}

func (v bindat) MarshalInto(data []byte) {
	_ = data[v.width()-1]
	b2int(len(v)).MarshalInto(data)
	copy(data[2:], []byte(v))
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

func (v bindat) width() int {
	return 2 + len(v)
}

// ----------------------------------------

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901011
type vbint uint

func (v vbint) WriteTo(w io.Writer) (int64, error) {
	return src(v).WriteTo(w)
}

func (v vbint) MarshalInto(data []byte) {
	_ = data[v.width()-1]
	var i int
	if v == 0 {
		return
	}
	for v > 0 {
		encodedByte := byte(v % 128)
		v = v / 128
		if v > 0 {
			encodedByte = encodedByte | 128
		}
		if i == len(data) {
			break
		}
		data[i] = encodedByte
		i++
		//fmt.Printf("%v %v %v\n", i, v, encodedByte)
	}
}

// MarshalBinary always returns nil error
func (v vbint) MarshalBinary() ([]byte, error) {
	data := make([]byte, v.width()) // max four
	v.MarshalInto(data)
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

func (v vbint) width() int {
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
func (v bits) Toggle(on bool, bit byte) {
	if on {
		v |= bits(bit)
		return
	}
	v ^= bits(bit)
}

// ----------------------------------------

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901008
type b2int uint16

func (v b2int) WriteTo(w io.Writer) (int64, error) {
	return src(v).WriteTo(w)
}

func (v b2int) MarshalInto(data []byte) {
	_ = data[1]
	binary.BigEndian.PutUint16(data, uint16(v))
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

func (v b2int) width() int { return 2 }

// ----------------------------------------

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901009
type b4int uint32

func (v b4int) WriteTo(w io.Writer) (int64, error) {
	return src(v).WriteTo(w)
}

func (v b4int) MarshalInto(data []byte) {
	_ = data[3]
	binary.BigEndian.PutUint32(data, uint32(v))
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

func (v b4int) width() int { return 4 }

// ---------------------------------------------------------------------
// Readers and writers
// ---------------------------------------------------------------------

// limitedReader is a reader with a known size. This is needed to
// calculate the remaining length of a control packet without loading
// everything into memory.
type limitedReader struct {
	src io.ReadSeeker

	// width is the number of bytes the above reader will ever read
	// before returning EOF. Similar to io.LimitedReader, though it's
	// not updated after each read.
	width int
}

// WriteTo copies the source to the given writer and then resets the
// src.
func (l *limitedReader) WriteTo(w io.Writer) (int64, error) {
	if l == nil {
		return 0, nil
	}
	n, err := io.Copy(w, l.src)
	l.src.Seek(0, io.SeekStart) // reset
	return n, err
}

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

// firstByte header flags
const (
	RETAIN byte = 0b0000_0001
	QoS0   byte = 0b0000_0000
	QoS1   byte = 0b0000_0010
	QoS2   byte = 0b0000_0100
	//QoS3 firstByte = 0b0000_0110   malformed!
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
