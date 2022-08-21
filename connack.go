package mqtt

func NewConnAck() *ConnAck {
	return &ConnAck{}
}

type ConnAck struct {
	fixed      Bits
	flags      Bits
	reasonCode wuint8

	sessionExpiryInterval wuint32
	receiveMax            wuint16
	maxQoS                wuint8 // 0 or 1, 2
	retainAvailable       wuint8
	maxPacketSize         wuint32
	assignedClientID      wstring
	topicAliasMax         wuint16
	reasonString          wstring

	userProp                         []property
	WildcardSubscriptionAvailable    wbool
	SubscriptionIdentifiersAvailable wbool
	SharedSubscriptionAvailable      wbool
	ServerKeepAlive                  wuint16
	ResponseInformation              wstring
	ServerReference                  wstring
	AuthenticationMethod             wstring
	AuthenticationData               bindata
}
