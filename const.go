package mqtt

// MQTT Packet property identifier codes

const (
	// maybe todo rename with prefix Id and create value types for
	// each property, e.g.  type ContentType u8str
	PayloadFormatIndicator Ident = 0x01
	MessageExpiryInterval  Ident = 0x02
	ContentType            Ident = 0x03

	ResponseTopic   Ident = 0x08
	CorrelationData Ident = 0x09

	SubscriptionID Ident = 0x0b

	SessionExpiryInterval Ident = 0x11
	AssignedClientIdent   Ident = 0x12
	ServerKeepAlive       Ident = 0x13

	AuthMethod          Ident = 0x15
	AuthData            Ident = 0x16
	RequestProblemInfo  Ident = 0x17
	WillDelayInterval   Ident = 0x18
	RequestResponseInfo Ident = 0x19
	ResponseInformation Ident = 0x1a

	ServerReference Ident = 0x1c
	ReasonString    Ident = 0x1f

	ReceiveMax           Ident = 0x21
	TopicAliasMax        Ident = 0x22
	TopicAlias           Ident = 0x23
	MaximumQoS           Ident = 0x24
	RetainAvailable      Ident = 0x25
	UserProperty         Ident = 0x26
	MaxPacketSize        Ident = 0x27
	WildcardSubAvailable Ident = 0x28
	SubIdentAvailable    Ident = 0x29
	SharedSubAvailable   Ident = 0x30
)

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

// FirstByte header flags
const (
	RETAIN byte = 0b0000_0001
	QoS1   byte = 0b0000_0010
	QoS2   byte = 0b0000_0100
	QoS3   byte = 0b0000_0110 // malformed!
	DUP    byte = 0b0000_1000
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
	MalformedPacket:                     "Malformed packet",
	ProtocolError:                       "Protocol error",
	ReceiveMaximumExceeded:              "Receive maximum exceeded",
	PacketTooLarge:                      "Packet too large",
	RetainNotSupported:                  "Retain not supported",
	QoSNotSupported:                     "QoS not supported",
	SharedSubscriptionsNotSupported:     "Shared subscriptions not supported",
	SubscriptionIdentifiersNotSupported: "Subscription identifiers not supported",
	WildcardSubscriptionsNotSupported:   "Wildcard subscriptions not supported",
}

// Name an empty slice for increased readability when fill methods are
// used to only calculate length.
var _LEN []byte
