package mq

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
func NewConnect() *Connect {
	return &Connect{
		fixed:           bits(CONNECT),
		protocolName:    mqtt5,
		protocolVersion: 5,
	}
}

type Connect struct {
	// Fields are kept hidden so
	// - we can optimize memory storage without affecting the API
	// - users don't have to handle dependencies between fields and flags

	// order is optimized for memory padding
	fixed           bits
	flags           bits
	protocolVersion wuint8
	keepAlive       wuint16
	receiveMax      wuint16

	sessionExpiryInterval wuint32
	maxPacketSize         wuint32

	willDelayInterval wuint32

	topicAliasMax       wuint16
	requestResponseInfo wbool
	requestProblemInfo  wbool

	protocolName wstring
	clientID     wstring
	UserProperties
	authMethod wstring
	authData   bindata

	username wstring
	password bindata

	will *Publish
}

// Connect fields are exposed using methods to simplify the type
// conversion.

// SetWill sets the will message and delay in seconds. The Server
// delays publishing the Client’s Will Message until the Will Delay
// Interval has passed or the Session ends, whichever happens first.
func (c *Connect) SetWill(p *Publish, delayInterval uint32) {
	c.will = p
	c.flags.toggle(WillFlag, true)
	c.flags.toggle(WillRetain, p.Retain())
	c.willDelayInterval = wuint32(delayInterval)
	c.setWillQoS(p.QoS())
}

func (c *Connect) Will() *Publish {
	return c.will
}

func (c *Connect) HasFlag(v byte) bool { return c.flags.Has(v) }

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

func (c *Connect) setWillQoS(v uint8) {
	c.flags &= bits(^(WillQoS2 | WillQoS1)) // reset
	c.flags.toggle(v<<3, v < 3)
}
func (c *Connect) willQoS() uint8 {
	return (uint8(c.flags) & (WillQoS2 | WillQoS1)) >> 3
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

func (c *Connect) appendWillProperty(p UserProp) {
	c.will.UserProperties = append(c.will.UserProperties, p)
}

func (c *Connect) SetAuthMethod(v string) { c.authMethod = wstring(v) }
func (c *Connect) AuthMethod() string     { return string(c.authMethod) }

func (c *Connect) SetAuthData(v []byte) { c.authData = v }
func (c *Connect) AuthData() []byte     { return c.authData }

func (c *Connect) SetUsername(v string) {
	c.username = wstring(v)
	if len(v) == 0 {
		c.username = nil
	}
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

	i += c.fixed.fill(b, i)      // firstByte header
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

	// using c.propertyMap is slow compared to direct field access
	i += c.receiveMax.fillProp(b, i, ReceiveMax)
	i += c.sessionExpiryInterval.fillProp(b, i, SessionExpiryInterval)
	i += c.maxPacketSize.fillProp(b, i, MaxPacketSize)
	i += c.topicAliasMax.fillProp(b, i, TopicAliasMax)
	i += c.requestResponseInfo.fillProp(b, i, RequestResponseInfo)
	i += c.requestProblemInfo.fillProp(b, i, RequestProblemInfo)
	i += c.authMethod.fillProp(b, i, AuthMethod)
	i += c.authData.fillProp(b, i, AuthData)

	// User properties, in the spec it's defined before authentication
	// method. Though order should not matter, placed here to mimic
	// pahos order.
	i += c.UserProperties.properties(b, i)
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
			i += c.will.UserProperties.properties(b, i)

			return i - n
		}

		i += vbint(properties(_LEN, 0)).fill(b, i)
		i += properties(b, i)
		i += c.will.topicName.fill(b, i)
		i += c.will.payload.fill(b, i)
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
	get(&c.keepAlive)
	buf.getAny(c.propertyMap(), c.AddUserProperty)

	// payload
	get(&c.clientID)
	if bits(c.flags).Has(WillFlag) {
		c.will = NewPublish()
		c.will.SetQoS(c.willQoS())
		buf.getAny(c.willPropertyMap(), c.appendWillProperty)
		get(&c.will.topicName)
		get(&c.will.payload)
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
		PayloadFormatIndicator: &c.will.payloadFormat,
		MessageExpiryInterval:  &c.will.messageExpiryInterval,
		ContentType:            &c.will.contentType,
		ResponseTopic:          &c.will.responseTopic,
		CorrelationData:        &c.will.correlationData,
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

// String returns a short string describing the connect packet.
func (c *Connect) String() string {
	return fmt.Sprintf("%s %s %s%v %s %s %v bytes",
		firstByte(c.fixed).String(), connectFlags(c.flags),
		c.protocolName,
		c.protocolVersion,
		c.ClientID(),
		time.Duration(c.keepAlive)*time.Second,
		c.fill(_LEN, 0),
	)
}

type connectFlags byte

// String returns flags represented with a letter.
// Improper flags are marked with '!' and unset are marked with '-'.
//
//	UsernameFlag  u
//	PasswordFlag  p
//	WillRetain    r
//	WillQoS       1, 2 or !
//	WillFlag      2
//	CleanStart    s
//	Reserved      !
func (c connectFlags) String() string {
	flags := bytes.Repeat([]byte("-"), 8)

	mark := func(i int, flag byte, v byte) {
		if !bits(c).Has(flag) {
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

// CONNECT flags used in Connect.HasFlag()
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
