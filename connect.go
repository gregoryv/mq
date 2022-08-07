package mqtt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

func (p *Connect) Fill(h FixedHeader, r *bytes.Reader) error {
	p.fixed = h

	// variable header (without properties)
	r.Read(p.variable)

	// properties
	propLen, _ := ParseVarInt(r)
	p.properties = make([]byte, propLen)
	r.Read(p.properties)

	// payload
	p.payload = make([]byte, r.Len())
	_, err := r.Read(p.payload)

	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func NewConnect() *Connect {
	p := &Connect{
		variable: make([]byte, 10), // fixed in lenght
	}
	p.SetProtocolName(protoName)
	p.SetProtocolVersion(version5)
	p.SetKeepAlive(10)
	return p
}

// Should we keep fields as []byte types as to make it easy to read
// and write, as that is the main purpose of the protocol?
//
// or do we keep them aligned with their intended value making it
// easier to e.g. dump and or convert, is there bytes encoder perhaps?

type Connect struct {
	// fixed header
	fixed []byte

	// variable header
	variable   []byte
	properties []byte

	// payload
	payload []byte
}

// 3.1.2.1 Protocol Name
func (p *Connect) ProtocolName() string  { return string(p.variable[:6]) }
func (p *Connect) ProtocolVersion() byte { return p.variable[6] }
func (p *Connect) Flags() byte           { return p.variable[7] }

func (p *Connect) SetProtocolName(v []byte)  { copy(p.variable[:6], v) }
func (p *Connect) SetProtocolVersion(v byte) { p.variable[6] = v }
func (p *Connect) SetFlags(f byte)           { p.variable[7] |= f }

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
	all = append(all, p.properties...)
	all = append(all, p.payload...)
	return all
}

func (p *Connect) HasFlag(f byte) bool {
	return p.Flags()&f == f
}

func (p *Connect) dump() string {
	flags := bytes.Repeat([]byte("-"), 8)
	for i, f := range connectFlagOrder {
		if p.HasFlag(f) {
			flags[i] = shortConnectFlags[f]
		}
	}

	parts := []string{
		p.FixedHeader().String(),
		p.ProtocolName(),
		string(flags),
	}
	return strings.Join(parts, " ")
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
	l := len(p.variable) + len(p.properties) + len(p.payload)
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
	WillQoS1,
	WillQoS2,
	WillFlag,
	CleanStart,
	Reserved,
}

var ErrIncomplete = fmt.Errorf("incomplete")

var ErrEmptyFixedHeader = fmt.Errorf("empty fixed header")
