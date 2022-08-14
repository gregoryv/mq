package mqtt

import (
	"bytes"
	"fmt"
	"io"
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
	willPayload       []byte
	payloadFormat     bool

	messageExpiryInterval uint32
	contentType           string
	responseTopic         string
	correlationData       []byte

	username string
	password []byte
}

// flags settings
func (c *Connect) SetWillRetain(v bool) { c.toggle(WillRetain, v) }
func (c *Connect) SetWillFlag(v bool)   { c.toggle(WillFlag, v) }
func (c *Connect) SetCleanStart(v bool) { c.toggle(CleanStart, v) }

func (c *Connect) SetProtocolVersion(v uint8) { c.protocolVersion = v }
func (c *Connect) SetProtocolName(v string)   { c.protocolName = v }
func (c *Connect) SetClientID(v string)       { c.clientID = v }
func (c *Connect) SetKeepAlive(v uint16)      { c.keepAlive = v }

func (c *Connect) SetWillQoS(v uint8)                { c.willQoS = v }
func (c *Connect) SetSessionExpiryInterval(v uint32) { c.sessionExpiryInterval = v }
func (c *Connect) SetReceiveMax(v uint16)            { c.receiveMax = v }
func (c *Connect) SetMaxPacketSize(v uint32)         { c.maxPacketSize = v }
func (c *Connect) SetTopicAliasMax(v uint16)         { c.topicAliasMax = v }
func (c *Connect) SetRequestResponseInfo(v bool)     { c.requestResponseInfo = v }
func (c *Connect) SetRequestProblemInfo(v bool)      { c.requestProblemInfo = v }
func (c *Connect) AddUserProp(v property)            { c.userProp = append(c.userProp, v) }
func (c *Connect) AddWillProp(v property)            { c.willProp = append(c.willProp, v) }
func (c *Connect) SetAuthMethod(v string)            { c.authMethod = v }
func (c *Connect) SetAuthData(v []byte)              { c.authData = v }

func (c *Connect) SetWillDelayInterval(v uint32) { c.willDelayInterval = v }
func (c *Connect) SetContentType(v string)       { c.contentType = v }
func (c *Connect) SetResponseTopic(v string)     { c.responseTopic = v }
func (c *Connect) SetCorrelationData(v []byte)   { c.correlationData = v }

func (c *Connect) SetUsername(v string) { c.username = v }
func (c *Connect) SetPassword(v []byte) { c.password = v }

func (c *Connect) WriteTo(w io.Writer) (int64, error) {
	var (
		// calculate full size of packet to make it as efficient as
		// possible and allocate one []byte for everything
		hl   = c.variableHeader(nil)
		pl   = c.payload(nil)
		rem  = hl + pl
		size = 1 + vbint(rem).width() + rem
		b    = make([]byte, size)
		i    int
	)

	// FirstByte header
	b[0] = c.fixed
	i++
	// remaining length
	vbint(rem).MarshalInto(b[i:])
	i += vbint(rem).width()

	// Variable header
	c.variableHeader(b[i:])
	i += hl

	// Packet payload
	c.payload(b[i:])

	n, err := w.Write(b)
	return int64(n), err
}

func (c *Connect) variableHeader(b []byte) int {
	var (
		i     int
		build = (b != nil)
	)
	// Protocol name
	if build {
		u8str(c.protocolName).MarshalInto(b)
	}
	i += u8str(c.protocolName).width()

	// Protocol version
	if build {
		b[i] = c.protocolVersion
	}
	i++

	// Flags
	if build {
		b[i] = c.flags
	}
	i++

	// Keep alive
	if build {
		b2int(c.keepAlive).MarshalInto(b[i:])
	}
	i += 2

	// Properties
	proplen := c.properties(nil)
	if build {
		vbint(proplen).MarshalInto(b[i:])
	}
	i += vbint(proplen).width()

	if build {
		c.properties(b[i:])
	}
	i += proplen

	return i
}

// properties returns length properties in wire format, if b is nil
// nothing is written, used to calculate length.
func (c *Connect) properties(b []byte) int {
	c.updateFlags()
	var (
		i     int
		build = (b != nil)
	)

	// Session expiry interval
	if v := c.sessionExpiryInterval; v > 0 {
		if build {
			b[i] = SessionExpiryInterval
			b4int(v).MarshalInto(b[i+1:])
		}
		i += 5
	}

	// Receive maximum
	if v := c.receiveMax; v > 0 {
		if build {
			b[i] = ReceiveMax
			b2int(v).MarshalInto(b[i+1:])
		}
		i += 3
	}

	// Maximum packet size
	if v := c.maxPacketSize; v > 0 {
		if build {
			b[i] = MaxPacketSize
			b4int(v).MarshalInto(b[i+1:])
		}
		i += 5
	}

	// Topic alias maximum
	if v := c.topicAliasMax; v > 0 {
		if build {
			b[i] = TopicAliasMax
			b2int(v).MarshalInto(b[i+1:])
		}
		i += 3
	}

	// Request response information
	if c.requestResponseInfo {
		if build {
			b[i] = RequestResponseInfo
			b[i+1] = byte(1)
		}
		i += 2
	}

	// Request problem information
	if c.requestProblemInfo {
		if build {
			b[i] = RequestProblemInfo
			b[i+1] = byte(1)
		}
		i += 2
	}

	// Authentication method
	if v := c.authMethod; len(v) > 0 {
		if build {
			b[i] = AuthMethod
			u8str(v).MarshalInto(b[i+1:])
		}
		i += 1 + u8str(v).width()
	}

	// Authentication data
	if v := c.authData; len(v) > 0 {
		if build {
			b[i] = AuthData
			bindat(v).MarshalInto(b[i+1:])
		}
		i += 1 + bindat(v).width()
	}

	// User properties, in the spec it's defined before authentication
	// method. Though order should not matter, placed here to mimic
	// pahos order.
	for _, prop := range c.userProp {
		if build {
			b[i] = UserProperty
			prop.MarshalInto(b[i+1:])
		}
		i += 1 + prop.width()
	}
	return i
}

func (c *Connect) payload(b []byte) int {
	var (
		i     int
		build = (b != nil)
	)

	if build {
		u8str(c.clientID).MarshalInto(b)
	}
	i += u8str(c.clientID).width()

	// will
	if bits(c.flags).Has(WillFlag) {
		willLen := c.will(nil)
		if build {
			vbint(willLen).MarshalInto(b[i:])
		}
		i += vbint(willLen).width()

		if build {
			c.will(b[i:])
		}
		i += willLen

		if build {
			u8str(c.willTopic).MarshalInto(b[i:])
		}
		i += u8str(c.willTopic).width()

		if build {
			copy(b[i:], c.willPayload)
		}
		i += len(c.willPayload)
	}

	// User Name
	if bits(c.flags).Has(UsernameFlag) {
		if build {
			u8str(c.username).MarshalInto(b[i:])
		}
		i += u8str(c.username).width()
	}
	// Password
	if bits(c.flags).Has(PasswordFlag) {
		if build {
			u8str(c.password).MarshalInto(b[i:])
		}
		i += u8str(c.password).width()
	}

	return i
}

func (c *Connect) will(b []byte) int {
	var (
		i     int
		build = (b != nil)
	)

	// Will Properties
	if v := c.willDelayInterval; v > 0 {
		if build {
			b[i] = WillDelayInterval
			b4int(v).MarshalInto(b[i+1:])
		}
		i += 5
	}

	if c.payloadFormat {
		if build {
			b[i] = PayloadFormatIndicator
			b[i+1] = byte(1)
		}
		i += 2
	}

	if v := c.messageExpiryInterval; v > 0 {
		if build {
			b[i] = MessageExpiryInterval
			b4int(v).MarshalInto(b[i+1:])
		}
		i += 5
	}

	if v := c.contentType; len(v) > 0 {
		if build {
			b[i] = ContentType
			u8str(v).MarshalInto(b[i+1:])
		}
		i += 1 + u8str(v).width()
	}

	if v := c.responseTopic; len(v) > 0 {
		if build {
			b[i] = ResponseTopic
			u8str(v).MarshalInto(b[i+1:])
		}
		i += 1 + u8str(v).width()
	}

	if v := c.correlationData; len(v) > 0 {
		if build {
			b[i] = CorrelationData
			copy(b[i+1:], v)
		}
		i += 1 + len(v)
	}

	for _, prop := range c.willProp {
		if build {
			b[i] = UserProperty
			prop.MarshalInto(b[i+1:])
		}
		i += 1 + prop.width()
	}

	return i
}

// Settings

func (c *Connect) String() string {
	return fmt.Sprintf("%s %s %s", c.clientID,
		FirstByte(c.fixed).String(), c.Flags(),
	)
}

func (c *Connect) Flags() connectFlags {
	c.updateFlags()
	return connectFlags(c.flags)
}

func (c *Connect) updateFlags() {
	c.toggle(UsernameFlag, len(c.username) > 0)
	c.toggle(PasswordFlag, len(c.password) > 0)
	c.flags &= ^(WillQoS1 | WillQoS2) // reset
	c.toggle(c.willQoS<<3, c.willQoS < 3)
}

func (c *Connect) toggle(flag byte, on bool) {
	if on {
		c.flags |= flag
		return
	}
	c.flags &= ^flag
}

// ---------------------------------------------------------------------
// 3.1.2.3 Connect Flags
// ---------------------------------------------------------------------

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
		if !c.Has(flag) {
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

func (c connectFlags) Has(f byte) bool { return bits(c).Has(f) }

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
