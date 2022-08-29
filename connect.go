package mqtt

import (
	"bytes"
	"fmt"
	"io"
	"time"
)

// If we want to be able to handle large packets each must implement
// io.ReaderFrom This allows a client decide if it should read in all
// the data in one slice and wrap it in a reader or not.

// The other direction is also important to be able to write out large
// packets without loading everything into memory each packet must
// implement io.WriterTo.

var mqtt5 = []byte("MQTT")

// NewConnect returns an empty MQTT v5 connect packet.
func NewConnect() Connect {
	return Connect{
		fixed:           Bits(CONNECT),
		protocolName:    mqtt5,
		protocolVersion: 5,
	}
}

type Connect struct {
	// Fields are kept hidden so
	// - we can optimize memory storage without affecting the API
	// - users don't have to handle dependencies between fields and flags

	// order is optimized for memory padding
	fixed           Bits
	flags           Bits
	protocolVersion wuint8
	willQoS         wuint8 // todo remove this one, part of flags
	keepAlive       wuint16
	receiveMax      wuint16 // 8

	sessionExpiryInterval wuint32
	maxPacketSize         wuint32 // 8

	willDelayInterval         wuint32
	willMessageExpiryInterval wuint32 // 8

	topicAliasMax       wuint16
	requestResponseInfo wbool
	requestProblemInfo  wbool
	willPayloadFormat   wbool

	protocolName wstring
	clientID     wstring
	userProp     []property
	willProp     []property
	authMethod   wstring
	authData     bindata

	willTopic       wstring
	willPayload     bindata
	willContentType wstring
	responseTopic   wstring
	correlationData bindata

	username wstring
	password bindata
}

// Connect fields are exposed using methods to simplify the type
// conversion.

func (c *Connect) Flags() Bits         { return c.flags }
func (c *Connect) HasFlag(v byte) bool { return c.flags.Has(v) }

func (c *Connect) SetWillRetain(v bool) {
	c.flags.toggle(WillRetain, v)
	c.flags.toggle(WillFlag, true)
}
func (c *Connect) WillRetain() bool {
	return c.HasFlag(WillRetain)
}

func (c *Connect) SetCleanStart(v bool) {
	c.flags.toggle(CleanStart, v)
}

func (c *Connect) SetProtocolVersion(v uint8) { c.protocolVersion = wuint8(v) }
func (c *Connect) ProtocolVersion() uint8     { return uint8(c.protocolVersion) }

func (c *Connect) SetProtocolName(v string) { c.protocolName = wstring(v) }
func (c *Connect) ProtocolName() string     { return string(c.protocolName) }

func (c *Connect) SetClientID(v string) { c.clientID = wstring(v) }
func (c *Connect) ClientID() string     { return string(c.clientID) }

func (c *Connect) SetKeepAlive(v uint16) { c.keepAlive = wuint16(v) }
func (c *Connect) KeepAlive() uint16     { return uint16(c.keepAlive) }

func (c *Connect) SetWillQoS(v uint8) {
	c.willQoS = wuint8(v)
	c.flags &= Bits(^(WillQoS2 | WillQoS1)) // reset
	c.flags.toggle(byte(c.willQoS<<3), v < 3)
}
func (c *Connect) WillQoS() uint8 {
	return uint8(c.willQoS)
}

func (c *Connect) SetSessionExpiryInterval(v uint32) {
	c.sessionExpiryInterval = wuint32(v)
}
func (c *Connect) SessionExpiryInterval() uint32 {
	return uint32(c.sessionExpiryInterval)
}

func (c *Connect) SetReceiveMax(v uint16) { c.receiveMax = wuint16(v) }
func (c *Connect) ReceiveMax() uint16     { return uint16(c.receiveMax) }

func (c *Connect) SetMaxPacketSize(v uint32) { c.maxPacketSize = wuint32(v) }
func (c *Connect) MaxPacketSize() uint32     { return uint32(c.maxPacketSize) }

// This value indicates the highest value that the Client will accept
// as a Topic Alias sent by the Server. The Client uses this value to
// limit the number of Topic Aliases that it is willing to hold on
// this Connection.
func (c *Connect) SetTopicAliasMax(v uint16) {
	c.topicAliasMax = wuint16(v)
}
func (c *Connect) TopicAliasMax() uint16 { return uint16(c.topicAliasMax) }

// The Client uses this value to request the Server to return Response
// Information in the CONNACK
func (c *Connect) SetRequestResponseInfo(v bool) {
	c.requestResponseInfo = wbool(v)
}
func (c *Connect) RequestResponseInfo() bool {
	return bool(c.requestResponseInfo)
}

// The Client uses this value to indicate whether the Reason String or
// User Properties are sent in the case of failures.
func (c *Connect) SetRequestProblemInfo(v bool) {
	c.requestProblemInfo = wbool(v)
}
func (c *Connect) RequestProblemInfo() bool {
	return bool(c.requestProblemInfo)
}

// AddUserProp adds a user property. The User Property is allowed to
// appear multiple times to represent multiple name, value pairs. The
// same name is allowed to appear more than once.
func (c *Connect) AddUserProp(key, val string) {
	c.AddUserProperty(property{key, val})
}
func (c *Connect) AddUserProperty(p property) {
	c.appendUserProperty(p)
}
func (c *Connect) appendUserProperty(p property) {
	c.userProp = append(c.userProp, p)
}

func (c *Connect) AddWillProp(key, val string) {
	c.AddWillProperty(property{key, val})
}
func (c *Connect) AddWillProperty(p property) {
	c.appendWillProperty(p)
	c.flags.toggle(WillFlag, true)
}
func (c *Connect) appendWillProperty(p property) {
	c.willProp = append(c.willProp, p)
}

func (c *Connect) SetAuthMethod(v string) { c.authMethod = wstring(v) }
func (c *Connect) AuthMethod() string     { return string(c.authMethod) }

func (c *Connect) SetAuthData(v []byte) { c.authData = v }
func (c *Connect) AuthData() []byte     { return c.authData }

// SetWillDelayInterval in seconds. The Server delays publishing the
// Client’s Will Message until the Will Delay Interval has passed or
// the Session ends, whichever happens first.
func (c *Connect) SetWillDelayInterval(v uint32) {
	c.willDelayInterval = wuint32(v)
	c.flags.toggle(WillFlag, true)
}
func (c *Connect) WillDelayInterval() uint32 {
	return uint32(c.willDelayInterval)
}

// the lifetime of the Will Message in seconds and is sent as the
// Publication Expiry Interval when the Server publishes the Will
// Message.
func (c *Connect) SetWillMessageExpiryInterval(v uint32) {
	c.willMessageExpiryInterval = wuint32(v)
	c.flags.toggle(WillFlag, true)
}
func (c *Connect) WillMessageExpiryInterval() uint32 {
	return uint32(c.willMessageExpiryInterval)
}

func (c *Connect) SetWillTopic(v string) {
	c.willTopic = wstring(v)
	c.flags.toggle(WillFlag, true)
}
func (c *Connect) WillTopic() string { return string(c.willTopic) }

// SetWillPayloadFormat, false indicates that the Will Message is
// unspecified bytes. True indicates that the Will Message is UTF-8
// Encoded Character Data.
func (c *Connect) SetWillPayloadFormat(v bool) {
	c.willPayloadFormat = wbool(v)
}
func (c *Connect) WillPayloadFormat() bool {
	return bool(c.willPayloadFormat)
}

func (c *Connect) SetWillPayload(v []byte) {
	c.willPayload = v
	c.flags.toggle(WillFlag, true)
}
func (c *Connect) WillPayload() []byte { return c.willPayload }

// The value of the Content Type is defined by the sending and
// receiving application, e.g. it may be a mime type like
// application/json.
func (c *Connect) SetWillContentType(v string) {
	c.willContentType = wstring(v)
	c.flags.toggle(WillFlag, true)
}
func (c *Connect) WillContentType() string { return string(c.willContentType) }

// SetResponseTopic a UTF-8 encoded string which is used as the topic
// name for a response message.
func (c *Connect) SetResponseTopic(v string) {
	c.responseTopic = wstring(v)
}
func (c *Connect) ResponseTopic() string {
	return string(c.responseTopic)
}

// The Correlation Data is used by the sender of the Request Message
// to identify which request the Response Message is for when it is
// received.
func (c *Connect) SetCorrelationData(v []byte) {
	c.correlationData = v
}
func (c *Connect) CorrelationData() []byte {
	return c.correlationData
}

func (c *Connect) SetUsername(v string) {
	c.username = wstring(v)
	c.flags.toggle(UsernameFlag, len(c.username) > 0)

}
func (c *Connect) Username() string { return string(c.username) }

func (c *Connect) SetPassword(v []byte) {
	c.password = v
	c.flags.toggle(PasswordFlag, len(c.password) > 0)
}
func (c *Connect) Password() []byte { return c.password }

// WriteTo writes this connect control packet in wire format to the
// given writer.
func (c *Connect) WriteTo(w io.Writer) (int64, error) {
	// allocate full size of entire packet
	b := make([]byte, c.fill(_LEN, 0))
	c.fill(b, 0)

	n, err := w.Write(b)
	return int64(n), err
}

func (c *Connect) fill(b []byte, i int) int {
	remainingLen := vbint(c.variableHeader(_LEN, 0) + c.payload(_LEN, 0))

	i += c.fixed.fill(b, i)      // FirstByte header
	i += remainingLen.fill(b, i) // remaining length
	i += c.variableHeader(b, i)  // variable header
	i += c.payload(b, i)         // payload

	return i
}

func (c *Connect) variableHeader(b []byte, i int) int {
	n := i

	i += c.protocolName.fill(b, i)               // Protocol name
	i += c.protocolVersion.fill(b, i)            // Protocol version
	i += c.flags.fill(b, i)                      // Flags
	i += c.keepAlive.fill(b, i)                  // Keep alive
	i += vbint(c.properties(_LEN, 0)).fill(b, i) // Properties len
	i += c.properties(b, i)                      // Properties

	return i - n
}

// properties returns length properties in wire format, if b is nil
// nothing is written, used to calculate length.
func (c *Connect) properties(b []byte, i int) int {
	n := i

	for id, v := range c.propertyMap() {
		i += v.fillProp(b, i, id)
	}

	// User properties, in the spec it's defined before authentication
	// method. Though order should not matter, placed here to mimic
	// pahos order.
	for j, _ := range c.userProp {
		i += c.userProp[j].fillProp(b, i, UserProperty)
	}
	return i - n
}

func (c *Connect) payload(b []byte, i int) int {
	n := i

	i += c.clientID.fill(b, i)

	if c.flags.Has(WillFlag) {
		// Inlined the will properties to bring it closer to the
		// payload, worked just as well with a Connect.will method.
		properties := func(b []byte, i int) int {
			n := i

			for id, v := range c.willPropertyMap() {
				i += v.fillProp(b, i, id)
			}
			for j, _ := range c.willProp {
				i += c.willProp[j].fillProp(b, i, UserProperty)
			}

			return i - n
		}

		i += vbint(properties(_LEN, 0)).fill(b, i)
		i += properties(b, i)
		i += c.willTopic.fill(b, i)   // topic
		i += c.willPayload.fill(b, i) // payload
	}

	if c.flags.Has(UsernameFlag) {
		i += c.username.fill(b, i)
	}
	if c.flags.Has(PasswordFlag) {
		i += c.password.fill(b, i)
	}

	return i - n
}

func (c *Connect) UnmarshalBinary(p []byte) error {
	// get guards against errors, it also advances the index
	buf := &buffer{data: p}
	get := buf.get

	// variable header
	get(&c.protocolName)
	get(&c.protocolVersion)
	get(&c.flags)
	c.willQoS = Bits(c.flags&Bits(WillQoS2|WillQoS1)) >> 3
	get(&c.keepAlive)
	buf.getAny(c.propertyMap(), c.appendUserProperty)

	// payload
	get(&c.clientID)
	if Bits(c.flags).Has(WillFlag) {
		buf.getAny(c.willPropertyMap(), c.appendWillProperty)
		get(&c.willTopic)
		get(&c.willPayload)
	}
	// username
	if c.flags.Has(UsernameFlag) {
		get(&c.username)
	}
	// password
	if c.flags.Has(PasswordFlag) {
		get(&c.password)
	}
	return buf.Err()
}
func (c *Connect) willPropertyMap() map[Ident]wireType {
	return map[Ident]wireType{
		WillDelayInterval:      &c.willDelayInterval,
		PayloadFormatIndicator: &c.willPayloadFormat,
		MessageExpiryInterval:  &c.willMessageExpiryInterval,
		ContentType:            &c.willContentType,
		ResponseTopic:          &c.responseTopic,
		CorrelationData:        &c.correlationData,
	}
}

func (c *Connect) propertyMap() map[Ident]wireType {
	return map[Ident]wireType{
		ReceiveMax:            &c.receiveMax,
		SessionExpiryInterval: &c.sessionExpiryInterval,
		MaxPacketSize:         &c.maxPacketSize,
		TopicAliasMax:         &c.topicAliasMax,
		RequestResponseInfo:   &c.requestResponseInfo,
		RequestProblemInfo:    &c.requestProblemInfo,
		AuthMethod:            &c.authMethod,
		AuthData:              &c.authData,
	}
}

func (c *Connect) String() string {
	return fmt.Sprintf("%s %s %s%v %s %v bytes",
		FirstByte(c.fixed).String(), connectFlags(c.Flags()),
		c.protocolName,
		c.protocolVersion,
		time.Duration(c.keepAlive)*time.Second,
		c.fill(_LEN, 0),
	)
}

type connectFlags byte

// String returns flags represented with a letter.
// Improper flags are marked with '!' and unset are marked with '-'.
//
//   UsernameFlag  u
//   PasswordFlag  p
//   WillRetain    r
//   WillQoS       1, 2 or !
//   WillFlag      2
//   CleanStart    s
//   Reserved      !
func (c connectFlags) String() string {
	flags := bytes.Repeat([]byte("-"), 8)

	mark := func(i int, flag byte, v byte) {
		if !Bits(c).Has(flag) {
			return
		}
		flags[i] = v
	}
	mark(0, UsernameFlag, 'u')
	mark(1, PasswordFlag, 'p')
	mark(2, WillRetain, 'r')
	mark(3, WillQoS2, '2')
	mark(4, WillQoS1, '1')
	mark(3, WillQoS1|WillQoS2, '!')
	mark(4, WillQoS1|WillQoS2, '!')
	mark(5, WillFlag, 'w')
	mark(6, CleanStart, 's')
	mark(7, Reserved, '!')

	return string(flags) // + fmt.Sprintf(" %08b", c)
}

// CONNECT flags
// todo hide these and use methods on Connect, e.g. Connect.IsCleanStart()
const (
	Reserved byte = 1 << iota
	CleanStart
	WillFlag
	WillQoS1
	WillQoS2
	WillRetain
	PasswordFlag
	UsernameFlag
)
