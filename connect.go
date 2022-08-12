package mqtt

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

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
	fixed           byte  // 1
	flags           byte  // 1
	protocolVersion uint8 // 1
	protocolName    string

	payload []byte
}

func (c *Connect) WriteTo(w io.Writer) (int64, error) {
	p, err := nexus.NewPrinter(w)

	// variable header
	p.Write([]byte{c.fixed})
	// size of the remaining things, we need to know this before

	proto, e := u8str(c.protocolName).MarshalBinary()
	*err = e
	p.Write(proto)
	p.Write([]byte{c.protocolVersion, c.flags})

	varhead, e := c.variableHeader()
	*err = e

	remlen := sumlen(varhead) + len(c.payload)
	rem, e := vbint(remlen).MarshalBinary()
	*err = e
	p.Write(rem)
	varhead.WriteTo(p)
	p.Write(c.payload)
	return p.Written, *err
}

func (p *Connect) String() string {
	var sb strings.Builder
	sb.WriteString(typeNames[p.fixed&0b1111_0000])
	if f := p.fixedFlags(bits(p.fixed)); len(f) > 0 {
		sb.WriteString(" ")
		sb.Write(f)
	}
	return sb.String()
}

func sumlen(b net.Buffers) int {
	var l int
	for _, v := range b {
		l += len(v)
	}
	return l
}

func (p *Connect) variableHeader() (net.Buffers, error) {
	buf := make(net.Buffers, 0)

	if p.Is(CONNECT) {
		namelen, _ := b2int(len(p.protocolName)).MarshalBinary()
		buf = append(buf, namelen)
		buf = append(buf, []byte(p.protocolName))
		buf = append(buf, []byte{p.protocolVersion, p.flags})

		return nil, fmt.Errorf(": todo")
	}
	return buf, nil
}

func (p *Connect) Is(v byte) bool {
	return p.fixed&0b1111_0000 == v
}

// UnmarshalBinary unmarshals a control packets remaining data. The
// header must be set before calling this func. len(data) is the fixed
// headers remainig length.
/*func (p *ControlPacket) UnmarshalBinary(data []byte) error {
	return fmt.Errorf(": todo")
}*/

func (p *Connect) fixedFlags(h bits) []byte {
	switch byte(h) & 0b1111_0000 {

	case UNDEFINED:
		str := fmt.Sprintf("%04b", h)
		return []byte(strings.ReplaceAll(str, "0", "-"))

	case PUBLISH:
		flags := []byte("---")
		if h.Has(DUP) {
			flags[0] = 'd'
		}
		switch {
		case h.Has(QoS1 | QoS2):
			flags[1] = '!' // malformed
		case h.Has(QoS1):
			flags[1] = '1'
		case h.Has(QoS2):
			flags[1] = '2'
		}
		if h.Has(RETAIN) {
			flags[2] = 'r'
		}
		return flags

	default:
		return nil
	}
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

type KeepAlive b2int

// ---------------------------------------------------------------------
// 3.1.2.11 CONNECT Properties
// ---------------------------------------------------------------------

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901047
type PropertyLen vbint

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901048
type SessionExpiryInterval FourByteInt

func (s SessionExpiryInterval) String() string {
	return s.Duration().String()
}

func (s SessionExpiryInterval) Duration() time.Duration {
	return time.Duration(s) * time.Second
}

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901049
type ReceiveMax b2int

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901050
type MaxPacketSize FourByteInt

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901051
type TopicAliasMax b2int

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901052
type RequestResponseInfo byte

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901053
type RequestProblemInfo byte

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901054
type UserProperty u8pair

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901055
type AuthMethod u8str

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901056
type AuthData BinaryData
