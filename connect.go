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

// NewConnect returns an empty MQTT v5 connect packet.
func NewConnect() *Connect {
	return &Connect{
		fixed:           CONNECT,
		protocolName:    "MQTT",
		protocolVersion: 5,
	}
}

type Connect struct {
	fixed           byte
	flags           byte
	protocolVersion uint8
	protocolName    string
	clientID        string
	keepAlive       uint16

	// properties
	willQoS               uint8
	sessionExpiryInterval uint32
	receiveMax            uint16
	maxPacketSize         uint32
	topicAliasMax         uint16
	requestResponseInfo   bool
	requestProblemInfo    bool
	userProp              []property
	willProp              []property
	authMethod            string
	authData              []byte

	willDelayInterval uint32
	willTopic         string
	willPayloadFormat bool
	willPayload       []byte

	messageExpiryInterval uint32
	willContentType       string
	responseTopic         string
	correlationData       []byte

	username string
	password []byte
}

// exposed fields
func (c *Connect) KeepAlive() uint16   { return c.keepAlive }
func (c *Connect) ClientID() string    { return c.clientID }
func (c *Connect) Username() string    { return c.username }
func (c *Connect) Password() []byte    { return c.password }
func (c *Connect) WillTopic() string   { return c.willTopic }
func (c *Connect) WillPayload() []byte { return c.willPayload }
func (c *Connect) Flags() Bits         { return Bits(c.flags) }
func (c *Connect) HasFlag(v byte) bool { return Bits(c.flags).Has(v) }

// flags settings
func (c *Connect) SetWillRetain(v bool) { c.toggle(WillRetain, v) }
func (c *Connect) SetWillFlag(v bool)   { c.toggle(WillFlag, v) }
func (c *Connect) SetCleanStart(v bool) { c.toggle(CleanStart, v) }

func (c *Connect) SetProtocolVersion(v uint8) { c.protocolVersion = v }
func (c *Connect) SetProtocolName(v string)   { c.protocolName = v }
func (c *Connect) SetClientID(v string)       { c.clientID = v }
func (c *Connect) SetKeepAlive(v uint16)      { c.keepAlive = v }

func (c *Connect) SetWillQoS(v uint8) {
	c.willQoS = v
	c.flags &= ^(WillQoS1 | WillQoS2) // reset
	c.toggle(c.willQoS<<3, c.willQoS < 3)
}
func (c *Connect) SetSessionExpiryInterval(v uint32) { c.sessionExpiryInterval = v }
func (c *Connect) SetReceiveMax(v uint16)            { c.receiveMax = v }
func (c *Connect) SetMaxPacketSize(v uint32)         { c.maxPacketSize = v }
func (c *Connect) SetTopicAliasMax(v uint16)         { c.topicAliasMax = v }
func (c *Connect) SetRequestResponseInfo(v bool)     { c.requestResponseInfo = v }
func (c *Connect) SetRequestProblemInfo(v bool)      { c.requestProblemInfo = v }

func (c *Connect) AddUserProp(key, val string) {
	c.userProp = append(c.userProp, property{key, val})
}

func (c *Connect) AddWillProp(v property) { c.willProp = append(c.willProp, v) }
func (c *Connect) SetAuthMethod(v string) { c.authMethod = v }
func (c *Connect) SetAuthData(v []byte)   { c.authData = v }

func (c *Connect) SetWillDelayInterval(v uint32) { c.willDelayInterval = v }
func (c *Connect) SetWillTopic(v string)         { c.willTopic = v }
func (c *Connect) SetWillPayloadFormat(v bool)   { c.willPayloadFormat = v }
func (c *Connect) SetWillPayload(v []byte)       { c.willPayload = v }

func (c *Connect) SetWillContentType(v string) { c.willContentType = v }
func (c *Connect) SetResponseTopic(v string)   { c.responseTopic = v }
func (c *Connect) SetCorrelationData(v []byte) { c.correlationData = v }

func (c *Connect) SetUsername(v string) {
	c.username = v
	c.toggle(UsernameFlag, len(c.username) > 0)
}
func (c *Connect) SetPassword(v []byte) {
	c.password = v
	c.toggle(PasswordFlag, len(c.password) > 0)
}

func (c *Connect) WriteTo(w io.Writer) (int64, error) {
	var (
		// calculate full size of packet to make it as efficient as
		// possible and allocate one []byte for everything
		rem  = c.variableHeader(nil, 0) + c.payload(nil, 0)
		size = 1 + vbint(rem).width() + rem
		b    = make([]byte, size)
		i    int
	)

	i += Bits(c.fixed).fill(b, i) // firstByte header
	i += vbint(rem).fill(b, i)    // remaining length
	i += c.variableHeader(b, i)   // Variable header
	_ = c.payload(b, i)           // Packet payload

	n, err := w.Write(b)
	return int64(n), err
}

func (c *Connect) variableHeader(b []byte, i int) int {
	n := i

	i += u8str(c.protocolName).fill(b, i)       // Protocol name
	i += Bits(c.protocolVersion).fill(b, i)     // Protocol version
	i += Bits(c.flags).fill(b, i)               // Flags
	i += b2int(c.keepAlive).fill(b, i)          // Keep alive
	i += vbint(c.properties(nil, 0)).fill(b, i) // Properties len
	i += c.properties(b, i)                     // Properties

	return i - n
}

// properties returns length properties in wire format, if b is nil
// nothing is written, used to calculate length.
func (c *Connect) properties(b []byte, i int) int {
	n := i

	// Session expiry interval
	if v := c.sessionExpiryInterval; v > 0 {
		i += Bits(SessionExpiryInterval).fill(b, i)
		i += b4int(v).fill(b, i)
	}

	// Receive maximum
	if v := c.receiveMax; v > 0 {
		i += Bits(ReceiveMax).fill(b, i)
		i += b2int(v).fill(b, i)
	}

	// Maximum packet size
	if v := c.maxPacketSize; v > 0 {
		i += Bits(MaxPacketSize).fill(b, i)
		i += b4int(v).fill(b, i)
	}

	// Topic alias maximum
	if v := c.topicAliasMax; v > 0 {
		i += Bits(TopicAliasMax).fill(b, i)
		i += b2int(v).fill(b, i)
	}

	// Request response information
	if c.requestResponseInfo {
		i += Bits(RequestResponseInfo).fill(b, i)
		i += Bits(1).fill(b, i)
	}

	// Request problem information
	if c.requestProblemInfo {
		i += Bits(RequestProblemInfo).fill(b, i)
		i += Bits(1).fill(b, i)
	}

	// Authentication method
	if v := c.authMethod; len(v) > 0 {
		i += Bits(AuthMethod).fill(b, i)
		i += u8str(v).fill(b, i)
	}

	// Authentication data
	if v := c.authData; len(v) > 0 {
		i += Bits(AuthData).fill(b, i)
		i += bindat(v).fill(b, i)
	}

	// User properties, in the spec it's defined before authentication
	// method. Though order should not matter, placed here to mimic
	// pahos order.
	for _, prop := range c.userProp {
		i += Bits(UserProperty).fill(b, i)
		i += prop.fill(b, i)
	}
	return i - n
}

func (c *Connect) payload(b []byte, i int) int {
	n := i

	i += u8str(c.clientID).fill(b, i)

	// will
	if Bits(c.flags).Has(WillFlag) {
		i += vbint(c.will(nil, 0)).fill(b, i)
		i += c.will(b, i)
		i += u8str(c.willTopic).fill(b, i)    // topic
		i += bindat(c.willPayload).fill(b, i) // payload
	}

	// User Name
	if Bits(c.flags).Has(UsernameFlag) {
		i += u8str(c.username).fill(b, i)
	}
	// Password
	if Bits(c.flags).Has(PasswordFlag) {
		i += u8str(c.password).fill(b, i)
	}

	return i - n
}

func (c *Connect) will(b []byte, i int) int {
	n := i

	// Will Properties
	if v := c.willDelayInterval; v > 0 {
		i += Bits(WillDelayInterval).fill(b, i)
		i += b4int(v).fill(b, i)
	}

	if c.willPayloadFormat {
		i += Bits(PayloadFormatIndicator).fill(b, i)
		i += Bits(1).fill(b, i)
	}

	if v := c.messageExpiryInterval; v > 0 {
		i += Bits(MessageExpiryInterval).fill(b, i)
		i += b4int(v).fill(b, i)
	}

	if v := c.willContentType; len(v) > 0 {
		i += Bits(ContentType).fill(b, i)
		i += u8str(v).fill(b, i)
	}

	if v := c.responseTopic; len(v) > 0 {
		i += Bits(ResponseTopic).fill(b, i)
		i += u8str(v).fill(b, i)
	}

	if v := c.correlationData; len(v) > 0 {
		i += Bits(CorrelationData).fill(b, i)
		i += bindat(v).fill(b, i)
	}

	for _, prop := range c.willProp {
		i += Bits(UserProperty).fill(b, i)
		i += prop.fill(b, i)
	}

	return i - n
}

func (c *Connect) String() string {
	return fmt.Sprintf("%s %s %s %s", c.clientID,
		firstByte(c.fixed).String(), connectFlags(c.Flags()),
		time.Duration(c.keepAlive)*time.Second,
	)
}

func (c *Connect) toggle(flag byte, on bool) {
	if on {
		c.flags |= flag
		return
	}
	c.flags &= ^flag
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
	flags := bytes.Repeat([]byte("-"), 7)

	mark := func(i int, flag byte, v byte) {
		if !Bits(c).Has(flag) {
			return
		}
		flags[i] = v
	}
	mark(0, UsernameFlag, 'u')
	mark(1, PasswordFlag, 'p')
	mark(2, WillRetain, 'r')
	mark(3, WillQoS1, '1')
	mark(3, WillQoS2, '2')
	mark(3, WillQoS1|WillQoS2, '!')
	mark(4, WillFlag, 'w')
	mark(5, CleanStart, 's')
	mark(6, Reserved, '!')

	return string(flags) // + fmt.Sprintf(" %08b", c)
}

// CONNECT flags
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

// MQTT Packet property identifier codes
const (
	PayloadFormatIndicator byte = 0x01
	MessageExpiryInterval  byte = 0x02
	ContentType            byte = 0x03

	ResponseTopic   byte = 0x08
	CorrelationData byte = 0x09

	SubIdent byte = 0x0b

	SessionExpiryInterval byte = 0x11
	AssignedClientIdent   byte = 0x12
	ServerKeepAlive       byte = 0x13

	AuthMethod          byte = 0x15
	AuthData            byte = 0x16
	RequestProblemInfo  byte = 0x17
	WillDelayInterval   byte = 0x18
	RequestResponseInfo byte = 0x19
	ResponseInformation byte = 0x1a

	ServerReference byte = 0x1c
	ReasonString    byte = 0x1f

	ReceiveMax           byte = 0x21
	TopicAliasMax        byte = 0x22
	TopicAlias           byte = 0x23
	MaximumQoS           byte = 0x24
	RetainAvailable      byte = 0x25
	UserProperty         byte = 0x26
	MaxPacketSize        byte = 0x27
	WildcardSubAvailable byte = 0x28
	SubIdentAvailable    byte = 0x29
	SharedSubAvailable   byte = 0x30
)
