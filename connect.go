package mqtt

import "bytes"

func NewConnect() *Connect {
	// 3.1.2 CONNECT Variable Header
	variable := make([]byte, 0)

	// 3.1.2.1 Protocol Name
	variable = append(variable, protoName...)

	// 3.1.2.2 Protocol Version (Level)
	variable = append(variable, version5)

	// 3.1.2.3 Connect Flags
	variable = append(variable, 0)

	// 3.1.2.10 Keep Alive

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
	fixed    []byte
	variable []byte
}

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
