package mqtt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

func NewConnect() *Connect {
	p := &Connect{protoName: make([]byte, 6)}
	p.SetProtocolName(protoName)
	p.SetProtocolVersion(version5)
	p.SetKeepAlive(10)
	p.makeFixedHeader()
	return p
}

type Connect struct {
	// fixed header
	fixed []byte

	// variable
	protoName []byte
	protoVer  byte

	flags      byte
	keepAlive  uint16
	properties []byte

	// payload
	payload []byte
}

func (p *Connect) Fill(h FixedHeader, rest []byte) error {
	p.fixed = h
	if len(rest) < 10 {
		return fmt.Errorf("Connect.Fill %w", ErrIncomplete)
	}
	p.SetProtocolName(rest[:6])
	p.SetProtocolVersion(rest[6])
	p.SetFlags(rest[7])

	alive, _ := binary.Uvarint(rest[7:9])
	p.SetKeepAlive(uint16(alive))

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

// 3.1.2.1 Protocol Name
func (p *Connect) SetProtocolName(v []byte)  { copy(p.protoName, v) }
func (p *Connect) SetProtocolVersion(v byte) { p.protoVer = v }
func (p *Connect) SetKeepAlive(sec uint16)   { p.keepAlive = sec }
func (p *Connect) SetFlags(f byte)           { p.flags = f }

var ErrIncomplete = fmt.Errorf("incomplete")

var ErrEmptyFixedHeader = fmt.Errorf("empty fixed header")

func (p *Connect) FixedHeader() FixedHeader {
	return FixedHeader(p.fixed)
}

func (p *Connect) Reader() *bytes.Reader {
	return bytes.NewReader(p.Bytes())
}

func (p *Connect) Bytes() []byte {
	all := make([]byte, 0)
	all = append(all, p.fixed...)
	all = append(all, p.protoName...)
	all = append(all, p.protoVer, p.flags)
	alive := make([]byte, 2)
	binary.BigEndian.PutUint16(alive, p.keepAlive)
	all = append(all, alive...)
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
