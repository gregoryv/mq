package mqtt

import (
	"bytes"
	"fmt"
	"io"
)

func NewConnect() *Connect {
	// 3.1.2 CONNECT Variable Header
	variable := make([]byte, 10)

	// 3.1.2.1 Protocol Name
	copy(variable, protoName)

	// 3.1.2.2 Protocol Version (Level)
	variable[6] = version5

	// 3.1.2.3 Connect Flags
	variable[7] = 0

	// 3.1.2.10 Keep Alive
	variable[8] = 0
	variable[9] = 10 // 10s

	// 3.1.2.11 CONNECT Properties

	variableLen := NewVarInt(uint(len(variable)))

	// fixed header
	h := make([]byte, 0, 5)
	h = append(h, CONNECT)
	h = append(h, variableLen...)

	return &Connect{
		fixed:    h,
		variable: variable,
	}
}

type Connect struct {
	// headers
	fixed      []byte
	variable   []byte // variable header and payload
	properties []byte
	payload    []byte
}

func (p *Connect) SetFlag(f byte) {
	p.variable[7] &= f
}

func (p *Connect) Fill(fixedHeader []byte, rest []byte) error {
	p.fixed = fixedHeader
	if len(rest) < 10 {
		return fmt.Errorf("Connect.Fill %w", ErrIncomplete)
	}
	p.variable = rest[:10] // fixed length
	propLen, err := ParseVarInt(bytes.NewReader(rest[10:]))
	if err != nil {
		return err
	}
	width := len(NewVarInt(propLen)) // maybe optimise
	p.payload = rest[10+width:]
	return nil
}

var ErrIncomplete = fmt.Errorf("incomplete")

// ReadFrom reads remaining variable header and payload.
// The fixed header must be set before calling ReadFrom.
func (p *Connect) ReadFrom(r io.Reader) (int64, error) {
	p.variable = make([]byte, p.FixedHeader().RemLen())
	n, err := r.Read(p.variable)

	return int64(n), err
}

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
	all = append(all, p.variable...)
	return all
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
