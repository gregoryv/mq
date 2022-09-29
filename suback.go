package mq

import (
	"fmt"
	"io"
)

func NewSubAck() SubAck {
	return SubAck{fixed: Bits(SUBACK)}
}

type SubAck struct {
	fixed    Bits
	packetID wuint16
	userProp []property

	reasonString wstring
	reasonCodes  []uint8
}

func (p *SubAck) String() string {
	return fmt.Sprintf("%s p%v %v bytes",
		firstByte(p.fixed).String(),
		p.packetID,
		p.width(),
	)
}
func (p *SubAck) SetPacketID(v uint16) { p.packetID = wuint16(v) }
func (p *SubAck) PacketID() uint16     { return uint16(p.packetID) }

func (p *SubAck) SetReasonString(v string) { p.reasonString = wstring(v) }
func (p *SubAck) ReasonString() string     { return string(p.reasonString) }

func (p *SubAck) AddReasonCode(v ReasonCode) {
	p.reasonCodes = append(p.reasonCodes, uint8(v))
}

func (p *SubAck) ReasonCodes() []uint8 { return p.reasonCodes }

func (p *SubAck) AddUserProp(key, val string) {
	p.AddUserProperty(property{key, val})
}
func (p *SubAck) AddUserProperty(prop property) {
	p.userProp = append(p.userProp, prop)
}

// ---------------------------------------- end settings

func (p *SubAck) WriteTo(w io.Writer) (int64, error) {
	b := make([]byte, p.width())
	p.fill(b, 0)
	n, err := w.Write(b)
	return int64(n), err
}

func (p *SubAck) width() int {
	return p.fill(_LEN, 0)
}

func (p *SubAck) fill(b []byte, i int) int {
	remainingLen := vbint(
		p.variableHeader(_LEN, 0) + p.payload(_LEN, 0),
	)
	i += p.fixed.fill(b, i)      // firstByte header
	i += remainingLen.fill(b, i) // remaining length
	i += p.variableHeader(b, i)
	i += p.payload(b, i)

	return i
}

func (p *SubAck) variableHeader(b []byte, i int) int {
	n := i
	i += p.packetID.fill(b, i)
	i += vbint(p.properties(_LEN, 0)).fill(b, i)
	i += p.properties(b, i)
	return i - n
}

func (p *SubAck) properties(b []byte, i int) int {
	n := i
	for id, v := range p.propertyMap() {
		i += v.fillProp(b, i, id)
	}
	for _, v := range p.userProp {
		i += v.fillProp(b, i, UserProperty)
	}
	return i - n
}

func (p *SubAck) payload(b []byte, i int) int {
	n := i
	for j, _ := range p.reasonCodes {
		i += wuint8(p.reasonCodes[j]).fill(b, i)
	}
	return i - n
}

func (p *SubAck) UnmarshalBinary(data []byte) error {
	b := &buffer{data: data}
	b.get(&p.packetID)
	b.getAny(p.propertyMap(), p.AddUserProperty)

	p.reasonCodes = make([]uint8, len(data)-b.i)

	for i, _ := range p.reasonCodes {
		var v wuint8
		b.get(&v)
		p.reasonCodes[i] = uint8(v)
	}
	return b.err
}

func (p *SubAck) propertyMap() map[Ident]wireType {
	return map[Ident]wireType{
		ReasonString: &p.reasonString,
	}
}
