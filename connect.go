package mqtt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

func (p *Connect) Fill(h FixedHeader, rest []byte) error {
	p.fixed = h
	n := copy(p.variable, rest)
	if n < 10 {
		return fmt.Errorf("variable header %w", ErrIncomplete)
	}

	// parse properties
	propLen, err := ParseVarInt(bytes.NewReader(rest[10:]))
	if err != nil && err != io.EOF {
		return err
	}

	width := len(NewVarInt(propLen)) // maybe optimise
	if propLen > 0 {
		p.properties = rest[10+width : 10+width+int(propLen)]
	}

	// rest is the payload
	if 10+int(propLen) <= len(rest) {
		return fmt.Errorf("TODO: %w", ErrIncomplete)
	}
	p.payload = rest[10+width+int(propLen):]
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
func (p *Connect) SetProtocolName(v []byte) { copy(p.variable[:6], v) }

func (p *Connect) SetProtocolVersion(v byte) { p.variable[6] = v }

func (p *Connect) SetFlags(f byte) { p.variable[7] = f }

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
	all := make([]byte, 0) // maybe optimise later to a known size
	all = append(all, p.fixed...)
	all = append(all, p.variable...)
	all = append(all, p.properties...)
	all = append(all, p.payload...)
	return all
}

func (p *Connect) makeFixedHeader() {
	h := make([]byte, 0, 5)
	h = append(h, CONNECT)
	h = append(h, p.variableLength()...)
	p.fixed = h
}

func (p *Connect) variableLength() []byte {
	l := 10 + len(p.properties) + len(p.payload)
	return NewVarInt(uint(l))
}

// 3.1.2.3 Connect Flags
const (
	Reserved byte = 1 << iota
	CleanStart
	WillFlag
	WillQoS
	WillRetain
	PasswordFlag
	UsernameFlag
)

var ErrIncomplete = fmt.Errorf("incomplete")

var ErrEmptyFixedHeader = fmt.Errorf("empty fixed header")
