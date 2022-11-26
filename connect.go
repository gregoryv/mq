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
func (p *Connect) SetWill(will *Publish, delayInterval uint32) {
	p.will = will
	p.flags.toggle(WillFlag, true)
	p.flags.toggle(WillRetain, will.Retain())
	p.willDelayInterval = wuint32(delayInterval)
	p.setWillQoS(will.QoS())
}

func (p *Connect) Will() *Publish {
	return p.will
}

func (p *Connect) HasFlag(v byte) bool { return p.flags.Has(v) }

func (p *Connect) SetCleanStart(v bool) { p.flags.toggle(CleanStart, v) }
func (p *Connect) CleanStart() bool     { return p.flags.Has(CleanStart) }

func (p *Connect) SetProtocolVersion(v uint8) { p.protocolVersion = wuint8(v) }
func (p *Connect) ProtocolVersion() uint8     { return uint8(p.protocolVersion) }

func (p *Connect) SetProtocolName(v string) { p.protocolName = wstring(v) }
func (p *Connect) ProtocolName() string     { return string(p.protocolName) }

func (p *Connect) SetClientID(v string) { p.clientID = wstring(v) }
func (p *Connect) ClientID() string     { return string(p.clientID) }

func (p *Connect) SetKeepAlive(v uint16) { p.keepAlive = wuint16(v) }
func (p *Connect) KeepAlive() uint16     { return uint16(p.keepAlive) }

func (p *Connect) setWillQoS(v uint8) {
	p.flags &= bits(^(WillQoS2 | WillQoS1)) // reset
	p.flags.toggle(v<<3, v < 3)
}
func (p *Connect) willQoS() uint8 {
	return (uint8(p.flags) & (WillQoS2 | WillQoS1)) >> 3
}

func (p *Connect) SetSessionExpiryInterval(v uint32) {
	p.sessionExpiryInterval = wuint32(v)
}
func (p *Connect) SessionExpiryInterval() uint32 {
	return uint32(p.sessionExpiryInterval)
}

func (p *Connect) SetReceiveMax(v uint16) { p.receiveMax = wuint16(v) }
func (p *Connect) ReceiveMax() uint16     { return uint16(p.receiveMax) }

func (p *Connect) SetMaxPacketSize(v uint32) { p.maxPacketSize = wuint32(v) }
func (p *Connect) MaxPacketSize() uint32     { return uint32(p.maxPacketSize) }

// This value indicates the highest value that the Client will accept
// as a Topic Alias sent by the Server. The Client uses this value to
// limit the number of Topic Aliases that it is willing to hold on
// this Connection.
func (p *Connect) SetTopicAliasMax(v uint16) {
	p.topicAliasMax = wuint16(v)
}
func (p *Connect) TopicAliasMax() uint16 { return uint16(p.topicAliasMax) }

// The Client uses this value to request the Server to return Response
// Information in the CONNACK
func (p *Connect) SetRequestResponseInfo(v bool) {
	p.requestResponseInfo = wbool(v)
}
func (p *Connect) RequestResponseInfo() bool {
	return bool(p.requestResponseInfo)
}

// The Client uses this value to indicate whether the ReasonString String or
// User Properties are sent in the case of failures.
func (p *Connect) SetRequestProblemInfo(v bool) {
	p.requestProblemInfo = wbool(v)
}
func (p *Connect) RequestProblemInfo() bool {
	return bool(p.requestProblemInfo)
}

func (p *Connect) appendWillProperty(prop UserProp) {
	p.will.UserProperties = append(p.will.UserProperties, prop)
}

func (p *Connect) SetAuthMethod(v string) { p.authMethod = wstring(v) }
func (p *Connect) AuthMethod() string     { return string(p.authMethod) }

func (p *Connect) SetAuthData(v []byte) { p.authData = v }
func (p *Connect) AuthData() []byte     { return p.authData }

func (p *Connect) SetUsername(v string) {
	p.username = wstring(v)
	if len(v) == 0 {
		p.username = nil
	}
	p.flags.toggle(UsernameFlag, len(p.username) > 0)

}
func (p *Connect) Username() string { return string(p.username) }

func (p *Connect) SetPassword(v []byte) {
	p.password = v
	p.flags.toggle(PasswordFlag, len(p.password) > 0)
}
func (p *Connect) Password() []byte { return p.password }

// String returns a short string describing the connect packet.
func (p *Connect) String() string {
	return fmt.Sprintf("%s %s %s%v %s %s %v bytes",
		firstByte(p.fixed).String(), connectFlags(p.flags),
		p.protocolName,
		p.protocolVersion,
		p.ClientID(),
		time.Duration(p.keepAlive)*time.Second,
		p.fill(_LEN, 0),
	)
}

func (p *Connect) dump(w io.Writer) {
	fmt.Fprintf(w, "AuthData: %v\n", p.AuthData())
	fmt.Fprintf(w, "AuthMethod: %v\n", p.AuthMethod())
	fmt.Fprintf(w, "CleanStart: %v\n", p.CleanStart())
	fmt.Fprintf(w, "ClientID: %v\n", p.ClientID())
	fmt.Fprintf(w, "KeepAlive: %v\n", p.KeepAlive())
	fmt.Fprintf(w, "MaxPacketSize: %v\n", p.MaxPacketSize())
	fmt.Fprintf(w, "Password: %q\n", stars(len(p.Password())))
	fmt.Fprintf(w, "ProtocolName: %v\n", p.ProtocolName())
	fmt.Fprintf(w, "ProtocolVersion: %v\n", p.ProtocolVersion())
	fmt.Fprintf(w, "ReceiveMax: %v\n", p.ReceiveMax())
	fmt.Fprintf(w, "RequestProblemInfo: %v\n", p.RequestProblemInfo())
	fmt.Fprintf(w, "RequestResponseInfo: %v\n", p.RequestResponseInfo())
	fmt.Fprintf(w, "SessionExpiryInterval: %v\n", p.SessionExpiryInterval())
	fmt.Fprintf(w, "TopicAliasMax: %v\n", p.TopicAliasMax())
	fmt.Fprintf(w, "Username: %v\n", stars(len(p.Username())))

	if p.will != nil {
		fmt.Fprintln(w, "Will")
		p.will.dump(w)
	}

	p.UserProperties.dump(w)
}

func stars(v int) string {
	if v == 0 {
		return ""
	}
	return "*********"
}

// WriteTo writes this connect control packet in wire format to the
// given writer.
func (p *Connect) WriteTo(w io.Writer) (int64, error) {
	// allocate full size of entire packet
	b := make([]byte, p.fill(_LEN, 0))
	p.fill(b, 0)

	n, err := w.Write(b)
	return int64(n), err
}

func (p *Connect) fill(b []byte, i int) int {
	remainingLen := vbint(p.variableHeader(_LEN, 0) + p.payload(_LEN, 0))

	i += p.fixed.fill(b, i)      // firstByte header
	i += remainingLen.fill(b, i) // remaining length
	i += p.variableHeader(b, i)  // variable header
	i += p.payload(b, i)         // payload

	return i
}

func (p *Connect) variableHeader(b []byte, i int) int {
	n := i

	i += p.protocolName.fill(b, i)               // Protocol name
	i += p.protocolVersion.fill(b, i)            // Protocol version
	i += p.flags.fill(b, i)                      // Flags
	i += p.keepAlive.fill(b, i)                  // Keep alive
	i += vbint(p.properties(_LEN, 0)).fill(b, i) // Properties len
	i += p.properties(b, i)                      // Properties

	return i - n
}

// properties returns length properties in wire format, if b is nil
// nothing is written, used to calculate length.
func (p *Connect) properties(b []byte, i int) int {
	n := i

	// using p.propertyMap is slow compared to direct field access
	i += p.receiveMax.fillProp(b, i, ReceiveMax)
	i += p.sessionExpiryInterval.fillProp(b, i, SessionExpiryInterval)
	i += p.maxPacketSize.fillProp(b, i, MaxPacketSize)
	i += p.topicAliasMax.fillProp(b, i, TopicAliasMax)
	i += p.requestResponseInfo.fillProp(b, i, RequestResponseInfo)
	i += p.requestProblemInfo.fillProp(b, i, RequestProblemInfo)
	i += p.authMethod.fillProp(b, i, AuthMethod)
	i += p.authData.fillProp(b, i, AuthData)

	// User properties, in the spec it's defined before authentication
	// method. Though order should not matter, placed here to mimic
	// pahos order.
	i += p.UserProperties.properties(b, i)
	return i - n
}

func (p *Connect) payload(b []byte, i int) int {
	n := i

	i += p.clientID.fill(b, i)

	if p.flags.Has(WillFlag) {
		// Inlined the will properties to bring it closer to the
		// payload, worked just as well with a Connect.will method.
		properties := func(b []byte, i int) int {
			n := i

			for id, v := range p.willPropertyMap() {
				i += v.fillProp(b, i, id)
			}
			i += p.will.UserProperties.properties(b, i)

			return i - n
		}

		i += vbint(properties(_LEN, 0)).fill(b, i)
		i += properties(b, i)
		i += p.will.topicName.fill(b, i)
		i += p.will.payload.fill(b, i)
	}

	if p.flags.Has(UsernameFlag) {
		i += p.username.fill(b, i)
	}
	if p.flags.Has(PasswordFlag) {
		i += p.password.fill(b, i)
	}

	return i - n
}

func (p *Connect) UnmarshalBinary(data []byte) error {
	// get guards against errors, it also advances the index
	buf := &buffer{data: data}
	get := buf.get

	// variable header
	get(&p.protocolName)
	get(&p.protocolVersion)
	get(&p.flags)
	get(&p.keepAlive)
	buf.getAny(p.propertyMap(), p.appendUserProperty)

	// payload
	get(&p.clientID)
	if bits(p.flags).Has(WillFlag) {
		p.will = NewPublish()
		p.will.SetQoS(p.willQoS())
		buf.getAny(p.willPropertyMap(), p.appendWillProperty)
		get(&p.will.topicName)
		get(&p.will.payload)
	}
	// username
	if p.flags.Has(UsernameFlag) {
		get(&p.username)
	}
	// password
	if p.flags.Has(PasswordFlag) {
		get(&p.password)
	}
	return buf.Err()
}
func (p *Connect) willPropertyMap() map[Ident]wireType {
	return map[Ident]wireType{
		WillDelayInterval:      &p.willDelayInterval,
		PayloadFormatIndicator: &p.will.payloadFormat,
		MessageExpiryInterval:  &p.will.messageExpiryInterval,
		ContentType:            &p.will.contentType,
		ResponseTopic:          &p.will.responseTopic,
		CorrelationData:        &p.will.correlationData,
	}
}

func (p *Connect) propertyMap() map[Ident]wireType {
	return map[Ident]wireType{
		ReceiveMax:            &p.receiveMax,
		SessionExpiryInterval: &p.sessionExpiryInterval,
		MaxPacketSize:         &p.maxPacketSize,
		TopicAliasMax:         &p.topicAliasMax,
		RequestResponseInfo:   &p.requestResponseInfo,
		RequestProblemInfo:    &p.requestProblemInfo,
		AuthMethod:            &p.authMethod,
		AuthData:              &p.authData,
	}
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
