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

// Reason Codes
//
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Handling_errors

type ReasonCode byte

const (
	Success                             ReasonCode = 0x00 // The Connection is accepted.
	UnspecifiedError                    ReasonCode = 0x80 // The Server does not wish to reveal the reason for the failure, or none of the other Reason Codes apply.
	MalformedPacket                     ReasonCode = 0x81 // Malformed Packet
	ProtocolError                       ReasonCode = 0x82 // Protocol Error
	ImplementationSpecificError         ReasonCode = 0x83 // The CONNECT is valid but is not accepted by this Server.
	UnsupportedProtocolVersion          ReasonCode = 0x84 // The Server does not support the version of the MQTT protocol requested by the Client.
	ClientIdentifierNotValid            ReasonCode = 0x85 // The Client Identifier is a valid string but is not allowed by the Server.
	BadUserNameOrPassword               ReasonCode = 0x86 // The Server does not accept the User Name or Password specified by the Client
	NotAuthorized                       ReasonCode = 0x87 // The Client is not authorized to connect.
	ServerUnavailable                   ReasonCode = 0x88 // The MQTT Server is not available.
	ServerBusy                          ReasonCode = 0x89 // The Server is busy. Try again later.
	Banned                              ReasonCode = 0x8A // This Client has been banned by administrative action. Contact the server administrator.
	BadAuthenticationMethod             ReasonCode = 0x8C // The authentication method is not supported or does not match the authentication method currently in use.
	TopicNameInvalid                    ReasonCode = 0x90 // The Will Topic Name is not malformed, but is not accepted by this Server.
	ReceiveMaximumExceeded              ReasonCode = 0x93 // Receive Maximum exceeded
	PacketTooLarge                      ReasonCode = 0x95 // The CONNECT packet exceeded the maximum permissible size.
	QuotaExceeded                       ReasonCode = 0x97 // An implementation or administrative imposed limit has been exceeded.
	PayloadFormatInvalid                ReasonCode = 0x99 // The Will Payload does not match the specified Payload Format Indicator.
	RetainNotSupported                  ReasonCode = 0x9A // Retain not supported
	QoSNotSupported                     ReasonCode = 0x9B // QoS not supported
	UseAnotherServer                    ReasonCode = 0x9C // The Client should temporarily use another server.
	ServerMoved                         ReasonCode = 0x9D // The Client should permanently use another server.
	SharedSubscriptionsNotSupported     ReasonCode = 0x9E // Shared Subscriptions not supported
	ConnectionRateExceeded              ReasonCode = 0x9F // The connection rate limit has been exceeded.
	SubscriptionIdentifiersNotSupported ReasonCode = 0xA1 // Subscription Identifiers not supported
	WildcardSubscriptionsNotSupported   ReasonCode = 0xA2 // Wildcard Subscriptions not supported
)

// Name an empty slice for increased readability when fill methods are
// used to only calculate length.
var _LEN []byte
