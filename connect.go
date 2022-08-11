package mqtt

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"
)

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

func (p *Connect) String() string {
	var sb strings.Builder
	sb.WriteString(typeNames[p.fixed&0b1111_0000])
	if f := p.fixedFlags(bits(p.fixed)); len(f) > 0 {
		sb.WriteString(" ")
		sb.Write(f)
	}
	return sb.String()
}

func (p *Connect) Buffers() (net.Buffers, error) {
	buf := make(net.Buffers, 0)

	varhead, err := p.variableHeader() // todo handle error
	if err != nil {
		return nil, err
	}

	// fixed header
	buf = append(buf, []byte{byte(p.fixed)})
	remlen := VarByteInt(sumlen(varhead) + len(p.payload))
	rem, _ := remlen.MarshalBinary() // todo handle error
	buf = append(buf, rem)
	buf = append(buf, varhead...)
	buf = append(buf, p.payload)

	return buf, nil
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
		namelen, _ := TwoByteInt(len(p.protocolName)).MarshalBinary()
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

type KeepAlive TwoByteInt

// ---------------------------------------------------------------------
// 3.1.2.11 CONNECT Properties
// ---------------------------------------------------------------------

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901047
type PropertyLen VarByteInt

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901048
type SessionExpiryInterval FourByteInt

func (s SessionExpiryInterval) String() string {
	return s.Duration().String()
}

func (s SessionExpiryInterval) Duration() time.Duration {
	return time.Duration(s) * time.Second
}

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901049
type ReceiveMax TwoByteInt

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901050
type MaxPacketSize FourByteInt

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901051
type TopicAliasMax TwoByteInt

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901052
type RequestResponseInfo byte

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901053
type RequestProblemInfo byte

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901054
type UserProperty UTF8StringPair

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901055
type AuthMethod UTF8String

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901056
type AuthData BinaryData
