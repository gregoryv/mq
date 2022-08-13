package mqtt

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

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
	// fields are ordered to minimize memory allocation
	fixed byte // 1
	flags byte // 1

	protocolVersion byte // 1
	protocolName    u8str

	clientID u8str

	prop connectProp
	// will *willProp todo

	willDelayInterval b4int
}

func (c *Connect) WriteTo(w io.Writer) (int64, error) {
	p, err := nexus.NewPrinter(w)

	// ----------------------------------------
	// variable header
	w.Write([]byte{c.fixed})
	vbint(c.width()).WriteTo(p)
	c.protocolName.WriteTo(p)
	w.Write([]byte{
		c.protocolVersion,
		c.flags,
	})

	// connect properties
	c.prop.WriteTo(p)

	// ----------------------------------------
	// payload
	// Client Identifier
	u8str(c.clientID).WriteTo(w)

	// Will Properties
	if bits(c.flags).Has(WillFlag) {
		c.willDelayInterval.WriteTo(p)
	}
	// Will Topic
	// Will Payload
	// User Name
	// Password

	return p.Written, *err
}

// width returns the remaining length
func (c *Connect) width() int {
	n := 10 // always there

	n += c.willPropertiesWidth()
	n += c.prop.width()
	return n
}

func (c *Connect) willPropertiesWidth() int {
	var n int
	if bits(c.flags).Has(WillFlag) {
		n += c.willDelayInterval.width()
	}
	return n
}

func (c *Connect) SetWillDelayInterval(v b4int) {
	c.flags = c.flags | WillFlag
	c.willDelayInterval = v
}

func (c *Connect) String() string {
	return fmt.Sprintf("%s %s", Fixed(c.fixed).String(), c.Flags())
}

func (c *Connect) check() error {
	return newMalformed(c, "", fmt.Errorf("todo"))
}

func (c *Connect) Flags() ConnectFlags {
	return ConnectFlags(c.flags)
}

// ----------------------------------------

type connectProp struct {
	sessionExpiryInterval uint32     // 0x11
	receiveMax            uint16     // 0x21
	maxPacketSize         uint32     // 0x27
	topicAliasMax         uint16     // 0x22
	requestResponseInfo   bool       // 0x19
	requestProblemInfo    bool       // 0x17
	userProperties        []property // 0x26
	authMethod            string     // 0x15
	authData              []byte     // 0x16
}

func (p *connectProp) width() int {
	n, _ := p.WriteTo(ioutil.Discard)
	return int(n)
}

func (p *connectProp) WriteTo(w io.Writer) (int64, error) {
	dst, err := nexus.NewPrinter(w)

	if p.sessionExpiryInterval > 0 {
		dst.Write([]byte{SessionExpiryInterval})
		b4int(p.sessionExpiryInterval).WriteTo(dst)
	}
	if p.receiveMax > 0 {
		b2int(p.receiveMax).WriteTo(dst)
	}
	if p.maxPacketSize > 0 {
		b2int(p.maxPacketSize).WriteTo(dst)
	}
	if p.maxPacketSize > 0 {
		b4int(p.maxPacketSize).WriteTo(dst)
	}
	return dst.Written, *err
}

func (p *connectProp) Set() {

}

// ---------------------------------------------------------------------
// 3.1.2.3 Connect Flags
// ---------------------------------------------------------------------

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

type ConnectFlags byte

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
func (c ConnectFlags) String() string {
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

func (c ConnectFlags) Has(f byte) bool { return bits(c).Has(f) }

// limitedReader is a reader with a known size. This is needed to
// calculate the remaining length of a control packet without loading
// everything into memory.
type limitedReader struct {
	src io.ReadSeeker

	// width is the number of bytes the above reader will ever read
	// before returning EOF. Similar to io.LimitedReader, though it's
	// not updated after each read.
	width int
}

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
