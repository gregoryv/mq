package mq

import (
	"fmt"
	"io"
)

func NewDisconnect() *Disconnect {
	return &Disconnect{fixed: Bits(DISCONNECT)}
}

type Disconnect struct {
	fixed Bits

	reasonCode wuint8
	userProp   []property
}

func (p *Disconnect) SetReasonCode(v ReasonCode) { p.reasonCode = wuint8(v) }
func (p *Disconnect) ReasonCode() ReasonCode     { return ReasonCode(p.reasonCode) }

func (p *Disconnect) String() string {
	return fmt.Sprintf("%s %v bytes",
		firstByte(p.fixed).String(),
		p.width(),
	)
}
func (p *Disconnect) AddUserProp(key, val string) {
	p.AddUserProperty(property{key, val})
}
func (p *Disconnect) AddUserProperty(prop property) {
	p.userProp = append(p.userProp, prop)
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

func (p *Disconnect) properties(b []byte, i int) int {
	n := i
	for _, v := range p.userProp {
		i += v.fillProp(b, i, UserProperty)
	}
	return i - n
}
func (p *Disconnect) UnmarshalBinary(data []byte) error {
	b := &buffer{data: data}
	b.get(&p.reasonCode)
	b.getAny(p.propertyMap(), p.AddUserProperty)
	return b.err
}

func (p *Disconnect) propertyMap() map[Ident]wireType {
	return map[Ident]wireType{}
}
