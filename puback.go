package mqtt

import (
	"fmt"
	"io"
)

func NewPubAck() PubAck {
	return PubAck{
		fixed: Bits(PUBACK),
	}
}

type PubAck struct {
	fixed Bits

	packetID   wuint16
	reasonCode wuint8
	reason     wstring
	userProp   []property
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
	p.appendUserProperty(prop)
}
func (p *PubAck) appendUserProperty(prop property) {
	p.userProp = append(p.userProp, prop)
}

// end settings
// ----------------------------------------

func (p *PubAck) String() string {
	return fmt.Sprintf("%s ",
		FirstByte(p.fixed).String(),
	)
}

func (p *PubAck) WriteTo(w io.Writer) (int64, error) {
	// allocate full size of entire packet
	b := make([]byte, p.fill(_LEN, 0))
	p.fill(b, 0)

	n, err := w.Write(b)
	return int64(n), err
}

func (p *PubAck) fill(b []byte, i int) int {
	remainingLen := vbint(p.variableHeader(_LEN, 0))

	i += p.fixed.fill(b, i)      // FirstByte header
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
	fill := func(id Ident, v wireType) {
		i += id.fill(b, i)
		i += v.fill(b, i)
	}
	if v := p.reason; len(v) > 0 {
		fill(ReasonString, &v)
	}
	for _, v := range p.userProp {
		fill(UserProperty, &v)
	}
	return i - n
}

func (p *PubAck) UnmarshalBinary(data []byte) error {
	buf := &buffer{data: data}
	get := buf.get

	get(&p.packetID)

	// no more data, see 3.4.2.1 PUBACK Reason Code
	if len(data) == buf.i {
		return buf.err
	}

	get(&p.reasonCode)
	fields := map[Ident]wireType{
		ReasonString: &p.reason,
	}
	buf.getAny(fields, p.appendUserProperty)

	return buf.err
}
