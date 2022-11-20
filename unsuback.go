package mq

import (
	"fmt"
	"io"
)

// UnsubAck and SubAck are exactly the same except for the fixed
// byte. Keep for now.

func NewUnsubAck() *UnsubAck {
	return &UnsubAck{fixed: bits(UNSUBACK)}
}

type UnsubAck struct {
	fixed    bits
	packetID wuint16
	UserProperties

	reasonString wstring
	reasonCodes  []uint8
}

func (p *UnsubAck) String() string {
	return fmt.Sprintf("%s p%v %v bytes",
		firstByte(p.fixed).String(),
		p.packetID,
		p.width(),
	)
}

func (p *UnsubAck) dump(w io.Writer) {
	fmt.Fprintf(w, "PacketID: %v\n", p.PacketID())
	fmt.Fprintf(w, "ReasonString: %v\n", p.ReasonString())
	fmt.Fprintf(w, "ReasonCodes: %v\n", p.ReasonCodes())
	p.UserProperties.dump(w)
}

func (p *UnsubAck) SetPacketID(v uint16) { p.packetID = wuint16(v) }
func (p *UnsubAck) PacketID() uint16     { return uint16(p.packetID) }

func (p *UnsubAck) SetReasonString(v string) { p.reasonString = wstring(v) }
func (p *UnsubAck) ReasonString() string     { return string(p.reasonString) }

func (p *UnsubAck) AddReasonCode(v ReasonCode) {
	p.reasonCodes = append(p.reasonCodes, uint8(v))
}
func (p *UnsubAck) ReasonCodes() []uint8 { return p.reasonCodes }

func (p *UnsubAck) WriteTo(w io.Writer) (int64, error) {
	b := make([]byte, p.width())
	p.fill(b, 0)
	n, err := w.Write(b)
	return int64(n), err
}

func (p *UnsubAck) width() int {
	return p.fill(_LEN, 0)
}

func (p *UnsubAck) fill(b []byte, i int) int {
	remainingLen := vbint(
		p.variableHeader(_LEN, 0) + p.payload(_LEN, 0),
	)
	i += p.fixed.fill(b, i)      // firstByte header
	i += remainingLen.fill(b, i) // remaining length
	i += p.variableHeader(b, i)
	i += p.payload(b, i)

	return i
}

func (p *UnsubAck) variableHeader(b []byte, i int) int {
	n := i
	i += p.packetID.fill(b, i)
	i += vbint(p.properties(_LEN, 0)).fill(b, i)
	i += p.properties(b, i)
	return i - n
}

func (p *UnsubAck) properties(b []byte, i int) int {
	n := i
	for id, v := range p.propertyMap() {
		i += v.fillProp(b, i, id)
	}
	i += p.UserProperties.properties(b, i)
	return i - n
}

func (p *UnsubAck) payload(b []byte, i int) int {
	n := i
	for j, _ := range p.reasonCodes {
		i += wuint8(p.reasonCodes[j]).fill(b, i)
	}
	return i - n
}

func (p *UnsubAck) UnmarshalBinary(data []byte) error {
	b := &buffer{data: data}
	b.get(&p.packetID)
	b.getAny(p.propertyMap(), p.appendUserProperty)

	p.reasonCodes = make([]uint8, len(data)-b.i)

	for i, _ := range p.reasonCodes {
		var v wuint8
		b.get(&v)
		p.reasonCodes[i] = uint8(v)
	}
	return b.err
}

func (p *UnsubAck) propertyMap() map[Ident]wireType {
	return map[Ident]wireType{
		ReasonString: &p.reasonString,
	}
}
