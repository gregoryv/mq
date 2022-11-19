package mq

type ReasonCode byte

//go:generate stringer -type ReasonCode
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
	UnspecifiedError                    ReasonCode = 0x80 // ConnAck, PubAck, PubRec, SubAck, UnsubAck, Disconnect
	MalformedPacket                     ReasonCode = 0x81 // ConnAck, Disconnect
	ProtocolError                       ReasonCode = 0x82 // ConnAck, Disconnect
	ImplementationSpecificError         ReasonCode = 0x83 // ConnAck, PubAck, PubRec, SubAck, UnsubAck, Disconnect
	UnsupportedProtocolVersion          ReasonCode = 0x84 // ConnAck
	ClientIdentifierNotValid            ReasonCode = 0x85 // ConnAck
	BadUserNameOrPassword               ReasonCode = 0x86 // ConnAck
	NotAuthorized                       ReasonCode = 0x87 // ConnAck, PubAck, PubRec, SubAck, UnsubAck, Disconnect
	ServerUnavailable                   ReasonCode = 0x88 // ConnAck
	ServerBusy                          ReasonCode = 0x89 // ConnAck, Disconnect
	Banned                              ReasonCode = 0x8A // ConnAck
	ServerShuttingDown                  ReasonCode = 0x8B // Disconnect
	BadAuthenticationMethod             ReasonCode = 0x8C // ConnAck, Disconnect
	KeepAliveTimeout                    ReasonCode = 0x8D // Disconnect
	SessionTakenOver                    ReasonCode = 0x8E // Disconnect
	TopicFilterInvalid                  ReasonCode = 0x8F // SubAck, UnsubAck, Disconnect
	TopicNameInvalid                    ReasonCode = 0x90 // ConnAck, PubAck, PubRec, Disconnect
	PacketIdentifierInUse               ReasonCode = 0x91 // PubAck, PubRec, SubAck, UnsubAck
	PacketIdentifierNotFound            ReasonCode = 0x92 // PubRel, PubComp
	ReceiveMaximumExceeded              ReasonCode = 0x93 // Disconnect
	TopicAliasInvalid                   ReasonCode = 0x94 // Disconnect
	PacketTooLarge                      ReasonCode = 0x95 // ConnAck, Disconnect
	MessageRateToHigh                   ReasonCode = 0x96 // Disconnect
	QuotaExceeded                       ReasonCode = 0x97 // ConnAck, PubAck, PubRec, SubAck, Disconnect
	AdministrativeAction                ReasonCode = 0x98 // Disconnect
	PayloadFormatInvalid                ReasonCode = 0x99 // ConnAck, PubAck, PubRec, Disconnect
	RetainNotSupported                  ReasonCode = 0x9A // ConnAck, Disconnect
	QoSNotSupported                     ReasonCode = 0x9B // ConnAck, Disconnect
	UseAnotherServer                    ReasonCode = 0x9C // ConnAck, Disconnect
	ServerMoved                         ReasonCode = 0x9D // ConnAck, Disconnect
	SharedSubscriptionsNotSupported     ReasonCode = 0x9E // SubAck, Disconnect
	ConnectionRateExceeded              ReasonCode = 0x9F // ConnAck, Disconnect
	MaximumConnectTime                  ReasonCode = 0xA0 // Disconnect
	SubscriptionIdentifiersNotSupported ReasonCode = 0xA1 // SubAck, Disconnect
	WildcardSubscriptionsNotSupported   ReasonCode = 0xA2 // SubAck, Disconnect
)
