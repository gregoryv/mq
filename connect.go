package mqtt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

func (p *Connect) parseProperties(r *bytes.Reader, propLen int) error {
	left := r.Len() - propLen + 1
	for left < r.Len() {
		ident, err := r.ReadByte()
		debug.Printf("ParseConnect ident:%v %v, left:%v", ident, err, r.Len())
		switch ident {
		case SessionExpiryInterval:
			v := make([]byte, 4)
			_, _ = r.Read(v)
			sec, _ := binary.Uvarint(v)
			p.SetSessionExpiryInterval(time.Duration(sec) * time.Second)
		default:
			return fmt.Errorf("unknown property 0x%02x", ident)
		}
	}
	return nil
}

func (p *Connect) propertyBytes() []byte {
	var all []byte
	// SessionExpiryInterval
	all = append(all, SessionExpiryInterval)
	v := make([]byte, 4)
	sec := uint32(p.sessionExpiryInterval.Seconds())
	binary.BigEndian.PutUint32(v, sec)
	// MaxSessionExpiryInterval
	all = append(all, v...)
	return all
}

const (
	SessionExpiryInterval byte = 0x11
	ReceiveMax            byte = 0x21
	MaxPacketSize         byte = 0x27
)

var onnectPropertyNames = map[byte]string{
	SessionExpiryInterval: "SessionExpiryInterval",
	ReceiveMax:            "ReceiveMax",
	MaxPacketSize:         "MaxPacketSize",
}

// WIP
// ----------------------------------------

// ParseConnect returns a new connect packet from the header and
// remaining bytes. h.RemLen must be equal to len(rem).
func ParseConnect(h FixedHeader, remaining []byte) (*Connect, error) {
	p := NewConnect()
	p.fixed = h

	r := bytes.NewReader(remaining)
	// variable header (without properties)
	r.Read(p.variable)

	// properties
	propLen, _ := ParseVarInt(r)

	if propLen > 0 {
		if err := p.parseProperties(r, int(propLen)); err != nil {
			return p, err
		}
	}

	// payload
	p.payload = make([]byte, r.Len())
	_, _ = r.Read(p.payload)
	return p, nil
}

func NewConnect() *Connect {
	p := &Connect{
		variable: make([]byte, 10, 10),
	}
	p.SetProtocolName(protoName)
	p.SetProtocolVersion(version5)
	p.SetKeepAlive(10)
	p.SetSessionExpiryInterval(59 * time.Second)
	p.SetReceiveMax(0) // means max 65,535 if not present
	p.SetMaxPacketSize(4096)
	return p
}

// Connect as defined in 3.1 CONNECT - Connection Request
//
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901033
type Connect struct {
	// fixed header
	fixed []byte

	// variable header
	variable   []byte
	properties []byte

	sessionExpiryInterval time.Duration
	receiveMax            int16
	maxPacketSize         int32

	// payload
	payload []byte
}

// 3.1.2.1 Protocol Name
func (p *Connect) ProtocolName() string  { return string(p.variable[2:6]) }
func (p *Connect) ProtocolVersion() byte { return p.variable[6] }
func (p *Connect) Flags() byte           { return p.variable[7] }

func (p *Connect) SetProtocolName(v []byte)  { copy(p.variable[:6], v) }
func (p *Connect) SetProtocolVersion(v byte) { p.variable[6] = v }

func (p *Connect) SessionExpiryInterval() time.Duration {
	return p.sessionExpiryInterval
}

func (p *Connect) SetSessionExpiryInterval(dur time.Duration) {
	p.sessionExpiryInterval = dur
}

func (p *Connect) SetReceiveMax(v int16) {
	p.receiveMax = v
}

func (p *Connect) SetMaxPacketSize(v int32) {
	p.maxPacketSize = v
}

// SetFlags replaces the current flags with f
func (p *Connect) SetFlags(f byte) { p.variable[7] |= f }

// SetFlag enables the given flags
func (p *Connect) WithFlags(f byte) *Connect {
	p.variable[7] = f
	return p
}

func (p *Connect) SetKeepAlive(sec uint16) {
	binary.BigEndian.PutUint16(p.variable[8:], sec)
}

func (p *Connect) FixedHeader() FixedHeader {
	p.makeFixedHeader()
	return FixedHeader(p.fixed)
}

func (p *Connect) Reader() *bytes.Reader {
	return bytes.NewReader(p.Bytes())
}

func (p *Connect) Bytes() []byte {
	p.makeFixedHeader()
	all := make([]byte, 0, p.size())
	all = append(all, p.fixed...)
	all = append(all, p.variable...)
	prop := p.propertyBytes()
	all = append(all, NewVarInt(uint(len(prop)))...)
	all = append(all, prop...)
	all = append(all, p.payload...)
	return all
}

func (p *Connect) HasFlag(f byte) bool {
	return p.Flags()&f == f
}

func (p *Connect) String() string {
	return fmt.Sprintf("%s %s%v %s %v",
		p.FixedHeader(),
		p.ProtocolName(),
		p.ProtocolVersion(),
		connectFlags(p.Flags()),
		p.SessionExpiryInterval(),
	)
}

func (p *Connect) size() int {
	return len(p.fixed) +
		len(p.variable) +
		len(p.properties) +
		len(p.payload)
}
func (p *Connect) makeFixedHeader() {
	h := make([]byte, 0, 5)
	h = append(h, CONNECT)
	h = append(h, p.variableLength()...)
	p.fixed = h
}

func (p *Connect) variableLength() []byte {
	l := len(p.variable) + len(p.propertyBytes()) + len(p.payload)
	return NewVarInt(uint(l))
}

// 3.1.2.3 Connect Flags
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

var shortConnectFlags = map[byte]byte{
	Reserved:     'X',
	CleanStart:   's',
	WillFlag:     'w',
	WillQoS1:     '1',
	WillQoS2:     '2',
	WillRetain:   'r',
	PasswordFlag: 'p',
	UsernameFlag: 'u',
}

var connectFlagOrder = []byte{
	UsernameFlag,
	PasswordFlag,
	WillRetain,
	'-', // QoS,
	WillFlag,
	CleanStart,
	Reserved,
}

type connectFlags byte

func (c connectFlags) String() string {
	flags := bytes.Repeat([]byte("-"), 7)
	for i, f := range connectFlagOrder {
		if c.Has(f) {
			flags[i] = shortConnectFlags[f]
		}
	}
	if c.Has(WillQoS1) {
		flags[3] = '1'
	}
	if c.Has(WillQoS2) {
		flags[3] = '2'
	}
	if c.Has(Reserved) {
		flags[6] = 'R'
	}
	return string(flags)
}

func (c connectFlags) Has(f byte) bool { return byte(c)&f == f }

var ErrIncomplete = fmt.Errorf("incomplete")

// ----------------------------------------

// connect properties
