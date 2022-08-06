package mqtt

import "bytes"

func NewConnect() *Connect {
	// 3.1.2 CONNECT Variable Header

	// 3.1.2.1 Protocol Name
	protoName := []byte{0, 4, 'M', 'Q', 'T', 'T'}
	l := len(protoName)

	// 3.1.2.2 Protocol Version

	// 3.1.2.3 Connect Flags

	// 3.1.2.10 Keep Alive

	// 3.1.2.11 CONNECT Properties

	// fixed header
	h := make([]byte, 0, 5)
	h = append(h, CONNECT)
	h = append(h, NewVarInt(uint(l))...)

	return &Connect{
		fixed: h,
	}
}

type Connect struct {
	// headers
	fixed []byte
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
	return all
}
