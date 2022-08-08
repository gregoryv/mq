/*
Package mqtt provides a MQTT v5.0 protocol implementation

This package contains the protocol parsing features whereas
subpackages contain client and server implementations.  The
specification is found at
https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html

*/
package mqtt

import (
	"github.com/gregoryv/mqtt/internal"
)

// static protocol name, 3.1.2.1 Protocol Name
var protoName = []byte{0, 4, 'M', 'Q', 'T', 'T'}

const version5 byte = 5

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

var flagNames = internal.Dict{
	DUP:    "DUP",
	QoS0:   "QoS0",
	QoS1:   "QoS1",
	QoS2:   "QoS2",
	RETAIN: "RETAIN",
}

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
