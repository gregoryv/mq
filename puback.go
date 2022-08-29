package mqtt

import (
	"fmt"
	"io"
)

func NewPubAck() PubAck {
	return PubAck{fixed: Bits(PUBACK)}
}

// A PubAck packet is the response to a Publish packets, depending on
// the fixed header it can be one of PUBACK, PUBREC, PUBREL or PUBCOMP
type PubAck struct {
	fixed Bits

	packetID   wuint16
	reasonCode wuint8
	reason     wstring
	userProp   []property
}

func (p *PubAck) String() string {
	s := fmt.Sprintf("%s %v %v bytes",
		firstByte(p.fixed).String(),
		p.packetID,
		p.width(),
	)
	if p.reasonCode > 0 {
		return fmt.Sprintf("%s %s %s",
			s,
			ReasonCode(p.reasonCode).String(),
			p.reason,
		)
	}
	return s
}

func (p *PubAck) SetPacketID(v uint16) { p.packetID = wuint16(v) }
func (p *PubAck) PacketID() uint16     { return uint16(p.packetID) }

func (p *PubAck) SetReasonCode(v ReasonCode) { p.reasonCode = wuint8(v) }
func (p *PubAck) ReasonCode() ReasonCode     { return ReasonCode(p.reasonCode) }

func (p *PubAck) SetReason(v string) { p.reason = wstring(v) }
func (p *PubAck) Reason() string     { return string(p.reason) }

func (p *PubAck) AddUserProp(key, val string) {
	p.AddUserProperty(property{key, val})
}
func (p *PubAck) AddUserProperty(prop property) {
	p.userProp = append(p.userProp, prop)
}

func (p *PubAck) WriteTo(w io.Writer) (int64, error) {
	b := make([]byte, p.fill(_LEN, 0))
	p.fill(b, 0)
	n, err := w.Write(b)
	return int64(n), err
}

func (p *PubAck) width() int {
	return p.fill(_LEN, 0)
}

func (p *PubAck) fill(b []byte, i int) int {
	remainingLen := vbint(p.variableHeader(_LEN, 0))
	i += p.fixed.fill(b, i)      // firstByte header
	i += remainingLen.fill(b, i) // remaining length
	i += p.variableHeader(b, i)
	return i
}
func (p *PubAck) variableHeader(b []byte, i int) int {
	n := i
	i += p.packetID.fill(b, i)
	i += p.reasonCode.fillOpt(b, i)

	propl := vbint(p.properties(_LEN, 0))
	if propl > 0 {
		i += propl.fill(b, i)   // Properties len
		i += p.properties(b, i) // Properties
	}
	return i - n
}

func (p *PubAck) properties(b []byte, i int) int {
	n := i
	for id, v := range p.propertyMap() {
		i += v.fillProp(b, i, id)
	}
	for _, v := range p.userProp {
		i += v.fillProp(b, i, UserProperty)
	}
	return i - n
}

func (p *PubAck) UnmarshalBinary(data []byte) error {
	b := &buffer{data: data}
	b.get(&p.packetID)
	// no more data, see 3.4.2.1 PUBACK Reason Code
	if len(data) == b.i {
		return b.err
	}
	b.get(&p.reasonCode)
	b.getAny(p.propertyMap(), p.AddUserProperty)
	return b.err
}

func (p *PubAck) propertyMap() map[Ident]wireType {
	return map[Ident]wireType{
		ReasonString: &p.reason,
	}
}
