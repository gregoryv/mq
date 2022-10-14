package mq

// 2.1.2 MQTT Control Packet type
//
// https://docs.oasis-open.org/mq/mq/v5.0/os/mq-v5.0-os.html#_MQTT_Control_Packet
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

// MQTT Packet property identifier codes
// Ident is the same as wuint16 but is used to name the identifier codes
type Ident uint8

const (
	PayloadFormatIndicator Ident = 0x01
	MessageExpiryInterval  Ident = 0x02
	ContentType            Ident = 0x03
	ResponseTopic          Ident = 0x08
	CorrelationData        Ident = 0x09
	SubscriptionID         Ident = 0x0b
	SessionExpiryInterval  Ident = 0x11
	AssignedClientID       Ident = 0x12
	ServerKeepAlive        Ident = 0x13
	AuthMethod             Ident = 0x15
	AuthData               Ident = 0x16
	RequestProblemInfo     Ident = 0x17
	WillDelayInterval      Ident = 0x18
	RequestResponseInfo    Ident = 0x19
	ResponseInformation    Ident = 0x1a
	ServerReference        Ident = 0x1c
	ReasonString           Ident = 0x1f
	ReceiveMax             Ident = 0x21
	TopicAliasMax          Ident = 0x22
	TopicAlias             Ident = 0x23
	MaxQoS                 Ident = 0x24
	RetainAvailable        Ident = 0x25
	UserProperty           Ident = 0x26
	MaxPacketSize          Ident = 0x27
	WildcardSubAvailable   Ident = 0x28
	SubIDsAvailable        Ident = 0x29
	SharedSubAvailable     Ident = 0x2a
)

const (
	maxUint16 = 1<<16 - 1
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
	QoS1   byte = 0b0000_0010
	QoS2   byte = 0b0000_0100
	QoS3   byte = 0b0000_0110 // malformed!
	DUP    byte = 0b0000_1000
)

// Reason Codes
//
// MQTT_Version_5_0.html#_Toc3901031

type ReasonCode byte

const (
	Success               ReasonCode = 0x00 // ConnAck, PubAck, PubRec, PubRel, PubComp, UnsubAck, Auth
	NormalDisconnect      ReasonCode = 0x00 // Disconnect
	GrantedQoS0           ReasonCode = 0x00 // SubAck
	GrantedQoS1           ReasonCode = 0x01 // SubAck
	GrantedQoS2           ReasonCode = 0x02 // SubAck
	DisconnectWithWill    ReasonCode = 0x04 // Disconnect
	NoMatchingSubscribers ReasonCode = 0x10 // PubAck, PubRec
	NoSubscriptionExisted ReasonCode = 0x11 // UnsubAck
	ContinueAuth          ReasonCode = 0x18 // Auth
	ReAuthenticate        ReasonCode = 0x19 // Auth

	// failures >= 0x80
	UnspecifiedError                    ReasonCode = 0x80 //
	MalformedPacket                     ReasonCode = 0x81 //
	ProtocolError                       ReasonCode = 0x82 //
	ImplementationSpecificError         ReasonCode = 0x83 //
	UnsupportedProtocolVersion          ReasonCode = 0x84 //
	ClientIdentifierNotValid            ReasonCode = 0x85 //
	BadUserNameOrPassword               ReasonCode = 0x86 //
	NotAuthorized                       ReasonCode = 0x87 //
	ServerUnavailable                   ReasonCode = 0x88 //
	ServerBusy                          ReasonCode = 0x89 //
	Banned                              ReasonCode = 0x8A //
	BadAuthenticationMethod             ReasonCode = 0x8C //
	TopicNameInvalid                    ReasonCode = 0x90 //
	PacketIdentifierInUse               ReasonCode = 0x91 // PubAck, PubRec, SubAck, UnsubAck
	PacketIdentifierNotFound            ReasonCode = 0x92 // PubRel, PubComp
	ReceiveMaximumExceeded              ReasonCode = 0x93 //
	TopicAliasInvalid                   ReasonCode = 0x94 // Disconnect
	PacketTooLarge                      ReasonCode = 0x95 //
	MessageRateToHigh                   ReasonCode = 0x96 // Disconnect
	QuotaExceeded                       ReasonCode = 0x97 //
	AdministrativeAction                ReasonCode = 0x98 // Disconnect
	PayloadFormatInvalid                ReasonCode = 0x99 //
	RetainNotSupported                  ReasonCode = 0x9A //
	QoSNotSupported                     ReasonCode = 0x9B //
	UseAnotherServer                    ReasonCode = 0x9C //
	ServerMoved                         ReasonCode = 0x9D //
	SharedSubscriptionsNotSupported     ReasonCode = 0x9E //
	ConnectionRateExceeded              ReasonCode = 0x9F //
	MaximumConnectTime                  ReasonCode = 0xA0 // Disconnect
	SubscriptionIdentifiersNotSupported ReasonCode = 0xA1 //
	WildcardSubscriptionsNotSupported   ReasonCode = 0xA2 //
)

// Name an empty slice for increased readability when fill methods are
// used to only calculate length.
var _LEN []byte

// Filter option, used in Subscribe
type FilterOption byte
type Opt = FilterOption

const (
	OptQoS1    Opt = 1
	OptQoS2    Opt = 2
	OptQoS3    Opt = 3 // malformed
	OptNL      Opt = 1 << 2
	OptRAP     Opt = 1 << 3
	OptRetain1 Opt = 1 << 4
	OptRetain2 Opt = 2 << 4
	OptRetain3 Opt = 3 << 4 // malformed
)
