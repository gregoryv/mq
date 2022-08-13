package mqtt

import (
	"bytes"
	"fmt"
	"io"

	"github.com/gregoryv/nexus"
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
	protocolVersion byte
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
	userProperties        []property
	willProperties        []property
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
	username              string
	password              []byte
}

func (c *Connect) WriteTo(dst io.Writer) (int64, error) {
	varhead := c.variableHeader()
	payload := c.payload()

	remainingLen := varhead.Len() + len(payload)

	w := nexus.NewWriter(dst)
	w.WriteBinary(
		c.fixed, vbint(remainingLen), // Fixed header
		varhead.Bytes(), // Variable header
		payload,         // Payload
	)
	return w.Done()
}

func (c *Connect) variableHeader() *bytes.Buffer {
	c.updateFlags()
	var buf bytes.Buffer
	conprop := c.properties()

	nexus.NewWriter(&buf).WriteBinary(
		u8str(c.protocolName),        // Protocol Name
		c.protocolVersion,            // Protocol Level
		c.flags,                      // Connect Flags
		b2int(c.keepAlive),           // Keep Alive
		vbint(len(conprop)), conprop, // Properties
	)
	return &buf
}

func (c *Connect) properties() []byte {
	var buf bytes.Buffer
	w := nexus.NewWriter(&buf)

	if v := c.sessionExpiryInterval; v > 0 {
		w.WriteBinary(SessionExpiryInterval, b4int(v))
	}

	if v := c.receiveMax; v > 0 {
		w.WriteBinary(ReceiveMax, b2int(v))
	}

	if v := c.maxPacketSize; v > 0 {
		w.WriteBinary(MaxPacketSize, b4int(v))
	}

	if v := c.topicAliasMax; v > 0 {
		w.WriteBinary(TopicAliasMax, b2int(v))
	}

	if c.requestResponseInfo {
		w.WriteBinary(RequestResponseInfo, byte(1))
	}

	if c.requestProblemInfo {
		w.WriteBinary(RequestProblemInfo, byte(1))
	}

	for _, prop := range c.userProperties {
		w.WriteBinary(UserProperty, prop)
	}

	if v := c.authMethod; len(v) > 0 {
		w.WriteBinary(AuthMethod, u8str(v))
	}

	if v := c.authData; len(v) > 0 {
		w.WriteBinary(AuthData, v)
	}

	fmt.Println("buf len", buf.Len())
	fmt.Println("propLen", c.propertiesLen())
	return buf.Bytes()
}

func (c *Connect) propertiesLen() int {
	var n int
	if v := c.sessionExpiryInterval; v > 0 {
		n += 5
	}

	if v := c.receiveMax; v > 0 {
		n += 3
	}

	if v := c.maxPacketSize; v > 0 {
		n += 5
	}

	if v := c.topicAliasMax; v > 0 {
		n += 3
	}

	if c.requestResponseInfo {
		n += 2
	}

	if c.requestProblemInfo {
		n += 2
	}

	for _, prop := range c.userProperties {
		n += 1
		n += prop.width()
	}

	if v := c.authMethod; len(v) > 0 {
		n += 1
		n += u8str(v).width()
	}

	if v := c.authData; len(v) > 0 {
		n += 1
		n += len(v)
	}

	return n
}

func (c *Connect) payload() []byte {
	var buf bytes.Buffer
	w := nexus.NewWriter(&buf)
	w.WriteBinary(
		u8str(c.clientID),
	)
	// will
	if bits(c.flags).Has(WillFlag) {
		willprop := c.will()
		w.WriteBinary(
			vbint(len(willprop)),
			willprop,
			u8str(c.willTopic),
			c.willPayload,
		)
	}
	// User Name
	if bits(c.flags).Has(UsernameFlag) {
		w.WriteBinary(u8str(c.username))
	}
	// Password
	if bits(c.flags).Has(PasswordFlag) {
		w.WriteBinary(u8str(c.password))
	}

	return buf.Bytes()
}

func (c *Connect) will() []byte {
	var buf bytes.Buffer
	w := nexus.NewWriter(&buf)

	// Will Properties
	if v := c.willDelayInterval; v > 0 {
		w.WriteBinary(WillDelayInterval, b4int(v))
	}

	if c.payloadFormat {
		w.WriteBinary(PayloadFormatIndicator, byte(1))
	}

	if v := c.messageExpiryInterval; v > 0 {
		w.WriteBinary(MessageExpiryInterval, b4int(v))
	}

	if v := c.contentType; len(v) > 0 {
		w.WriteBinary(ContentType, u8str(v))
	}

	if v := c.responseTopic; len(v) > 0 {
		w.WriteBinary(ResponseTopic, u8str(v))
	}

	if v := c.correlationData; len(v) > 0 {
		w.WriteBinary(CorrelationData, v)
	}

	for _, prop := range c.willProperties {
		w.WriteBinary(UserProperty, prop)
	}

	return buf.Bytes()
}

// Settings

func (c *Connect) SetSessionExpiryInterval(v uint32) { c.sessionExpiryInterval = v }
func (c *Connect) SetReceiveMax(v uint16)            { c.receiveMax = v }
func (c *Connect) SetMaxPacketSize(v uint32)         { c.maxPacketSize = v }
func (c *Connect) SetTopicAliasMax(v uint16)         { c.topicAliasMax = v }
func (c *Connect) SetRequestResponseInfo(v bool)     { c.requestResponseInfo = v }
func (c *Connect) SetClientID(v string)              { c.clientID = v }
func (c *Connect) SetKeepAlive(v uint16)             { c.keepAlive = v }
func (c *Connect) SetWillRetain(v bool)              { c.toggle(WillRetain, v) }
func (c *Connect) SetWillFlag(v bool)                { c.toggle(WillFlag, v) }
func (c *Connect) SetCleanStart(v bool)              { c.toggle(CleanStart, v) }
func (c *Connect) SetUsername(v string)              { c.username = v }
func (c *Connect) SetPassword(v []byte)              { c.password = v }

// SetWillQoS, valid values are 0, 1 or 2
func (c *Connect) SetWillQoS(v uint8) { c.willQoS = v }

func (c *Connect) String() string {
	return fmt.Sprintf("%s %s %s", c.clientID,
		Fixed(c.fixed).String(), c.Flags(),
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
