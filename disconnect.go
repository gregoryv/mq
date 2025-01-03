package mq

import (
	"fmt"
	"io"
)

// NewDisconnect returns a disconnect packet with reason code
// NormalDisconnect.
func NewDisconnect() *Disconnect {
	return &Disconnect{fixed: bits(DISCONNECT)}
}

type Disconnect struct {
	fixed bits

	reasonCode wuint8
	UserProperties
}

func (p *Disconnect) SetReasonCode(v ReasonCode) { p.reasonCode = wuint8(v) }
func (p *Disconnect) ReasonCode() ReasonCode     { return ReasonCode(p.reasonCode) }

func (p *Disconnect) String() string {
	return withReason(p, fmt.Sprintf("%s %v bytes",
		firstByte(p.fixed).String(),
		p.width(),
	))
}

func (p *Disconnect) dump(w io.Writer) {
	fmt.Fprintf(w, "ReasonCode: %v\n", p.ReasonCode())
	p.UserProperties.dump(w)
}

func (p *Disconnect) WriteTo(w io.Writer) (int64, error) {
	b := make([]byte, p.width())
	p.fill(b, 0)
	n, err := w.Write(b)
	return int64(n), err
}

func (p *Disconnect) width() int {
	return p.fill(_LEN, 0)
}

func (p *Disconnect) fill(b []byte, i int) int {
	remainingLen := vbint(p.variableHeader(_LEN, 0))
	i += p.fixed.fill(b, i)      // firstByte header
	i += remainingLen.fill(b, i) // remaining length
	i += p.variableHeader(b, i)

	return i
}

func (p *Disconnect) variableHeader(b []byte, i int) int {
	n := i
	proplen := p.properties(_LEN, 0)
	if p.reasonCode == 0 && proplen == 0 {
		return 0
	}
	i += p.reasonCode.fill(b, i)
	i += vbint(proplen).fill(b, i)
	i += p.properties(b, i)
	return i - n
}

func (p *Disconnect) UnmarshalBinary(data []byte) error {
	b := &buffer{data: data}
	b.get(&p.reasonCode)
	b.getAny(p.propertyMap(), p.appendUserProperty)
	return b.err
}

func (p *Disconnect) propertyMap() map[Ident]func() wireType {
	return map[Ident]func() wireType{}
}
