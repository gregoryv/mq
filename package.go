/*
Package mqtt provides a MQTT v5.0 protocol implementation

The specification is found at
https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html

*/
package mqtt

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// firstByte represents the first byte in a control packet.
type firstByte byte

// String returns a readable string TYPEFLAGS, e.g. PUBLISH d1-r
func (f firstByte) String() string {
	var sb strings.Builder
	sb.WriteString(typeNames[byte(f)&0b1111_0000])
	sb.WriteString(" ")
	flags := []byte("----")
	if Bits(f).Has(DUP) {
		flags[0] = 'd'
	}
	switch {
	case Bits(f).Has(QoS1 | QoS2):
		flags[1] = '!' // malformed
		flags[2] = '!' // malformed
	case Bits(f).Has(QoS1):
		flags[2] = '1'
	case Bits(f).Has(QoS2):
		flags[1] = '2'
	}
	if Bits(f).Has(RETAIN) {
		flags[3] = 'r'
	}
	sb.Write(flags)
	return sb.String()
}

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901013
type property [2]string

func (v property) fill(data []byte, i int) int {
	i += u8str(v[0]).fill(data, i)
	_ = u8str(v[1]).fill(data, i)
	return v.width()
}

func (v *property) UnmarshalBinary(data []byte) error {
	var key u8str
	if err := key.UnmarshalBinary(data); err != nil {
		return unmarshalErr(v, "key", err.(*Malformed))
	}
	v[0] = string(key)

	i := len(v[0]) + 2
	var val u8str
	if err := val.UnmarshalBinary(data[i:]); err != nil {
		return unmarshalErr(v, "value", err.(*Malformed))
	}
	v[1] = string(val)
	return nil
}
func (v property) String() string {
	return fmt.Sprintf("%s:%s", v[0], v[1])
}
func (v property) width() int {
	return u8str(v[0]).width() + u8str(v[1]).width()
}

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901010
type u8str string

func (v u8str) fill(data []byte, i int) int {
	return bindata(v).fill(data, i)
}

func (v *u8str) UnmarshalBinary(data []byte) error {
	var b bindata
	if err := b.UnmarshalBinary(data); err != nil {
		return unmarshalErr(v, "", err.(*Malformed))
	}
	*v = u8str(b)
	return nil
}

func (v u8str) width() int {
	return bindata(v).width()
}

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901012
type bindata []byte

func (v bindata) fill(data []byte, i int) int {
	if len(data) >= i+v.width() {
		i += b2int(len(v)).fill(data, i)
		copy(data[i:], []byte(v))
	}
	return v.width()
}

func (v *bindata) UnmarshalBinary(data []byte) error {
	var length b2int
	_ = length.UnmarshalBinary(data)
	if len(data) < int(length)+2 {
		return unmarshalErr(v, "", "missing data")
	}
	*v = make([]byte, length)
	copy(*v, data[2:length+2])
	return nil
}

func (v bindata) width() int {
	return 2 + len(v)
}

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901011
type vbint uint

func (v vbint) fill(data []byte, i int) int {
	x := v
	n := i
	for {
		encodedByte := byte(x % 128)
		x = x / 128
		if x > 0 {
			encodedByte = encodedByte | 128
		}
		if i < len(data) {
			data[i] = encodedByte
		}
		i++
		if x == 0 {
			break
		}
	}
	return i - n
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

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901007
type Bits byte

func (v Bits) Has(b byte) bool { return byte(v)&b == b }
func (v Bits) fill(data []byte, i int) int {
	if len(data) >= i+1 {
		data[i] = byte(v)
	}
	return 1
}

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901008
type b2int uint16

func (v b2int) fill(data []byte, i int) int {
	if len(data) >= i+2 {
		binary.BigEndian.PutUint16(data[i:], uint16(v))
	}
	return 2
}

func (v *b2int) UnmarshalBinary(data []byte) error {
	*v = b2int(binary.BigEndian.Uint16(data))
	return nil
}

func (v b2int) width() int { return 2 }

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901009
type b4int uint32

func (v b4int) fill(data []byte, i int) int {
	if len(data) >= i+v.width() {
		binary.BigEndian.PutUint32(data[i:], uint32(v))
	}
	return v.width()
}

func (v *b4int) UnmarshalBinary(data []byte) error {
	*v = b4int(binary.BigEndian.Uint32(data))
	return nil
}

func (v b4int) width() int { return 4 }

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
