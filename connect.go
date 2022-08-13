package mqtt

import (
	"bytes"
	"fmt"
	"io"

	"github.com/gregoryv/nexus"
)

func connect() {

}

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
	sessionExpiryInterval uint32
	receiveMax            uint16
	maxPacketSize         uint32
	topicAliasMax         uint16
	requestResponseInfo   bool
	requestProblemInfo    bool
	userProperties        []property
	authMethod            string
	authData              []byte

	willDelayInterval uint32
	willTopic         string
	willPayload       string

	messageExpiryInterval uint32
	contentType           string
	responseTopic         string
	correlationData       []byte
	username              string
	password              []byte
}

func (c *Connect) WriteTo(w io.Writer) (int64, error) {
	{
		w := nexus.NewWriter(w)

		conprop := c.properties()

		w.WriteBinary(
			// Fixed header
			c.fixed,
			vbint(c.width()),

			// Variable header
			c.protocolVersion,     // Protocol Name
			u8str(c.protocolName), // Protocol Level
			c.flags,               // Connect Flags
			b2int(c.keepAlive),    // Keep Alive

			// Properties
			vbint(len(conprop)),
			conprop,

			// Payload
			u8str(c.clientID), // Client Identifier
		)
		// will
		if bits(c.flags).Has(WillFlag) {
			willprop := c.will()
			w.WriteBinary(
				vbint(len(willprop)),
				willprop,
			)

		}

		// User Name
		// Password

		return w.Done()
	}
}

// width returns the remaining length
func (c *Connect) width() int {
	n := 10 // always there
	return n
}

func (c *Connect) properties() []byte {
	var width int // todo
	p := make([]byte, width)
	return p
}

func (c *Connect) will() []byte {
	var buf bytes.Buffer
	w := nexus.NewWriter(&buf)
	w.WriteBinary(
	// Will Properties
	// Will Topic
	// Will Payload
	)
	return buf.Bytes()
}

func (c *Connect) String() string {
	return fmt.Sprintf("%s %s", Fixed(c.fixed).String(), c.Flags())
}

func (c *Connect) check() error {
	return newMalformed(c, "", fmt.Errorf("todo"))
}

func (c *Connect) Flags() connectFlags {
	return connectFlags(c.flags)
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

	return string(flags)
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
	PayloadFormat         byte = 1
	MessageExpiryInterval byte = 2
	ResponseTopic         byte = 8
	CorrelationData       byte = 9
	SubIdent              byte = 11
	SessionExpiryInterval byte = 17
	AssignedClientIdent   byte = 18
	ServerKeepAlive       byte = 19
	AuthMethod            byte = 21
	AuthData              byte = 22
	RequestProblemInfo    byte = 23
	WillDelayInterval     byte = 24
	RequestResponseInfo   byte = 25
	ResponseInformation   byte = 26
	ServerReference       byte = 28
	ReasonString          byte = 31
	ReceiveMax            byte = 33
	TopicAliasMax         byte = 34
	TopicAlias            byte = 35
	MaximumQoS            byte = 36
	RetainAvailable       byte = 37
	UserProperty          byte = 38
	MaxPacketSize         byte = 39
	WildcardSubAvailable  byte = 40
	SubIdentAvailable     byte = 41
	SharedSubAvailable    byte = 42
)
