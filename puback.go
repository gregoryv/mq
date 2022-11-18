package mq

import (
	"fmt"
	"io"
)

// NewPubAck returns control packet with type PUBACK
func NewPubAck() *PubAck {
	return &PubAck{fixed: bits(PUBACK)}
}

// A PubAck packet is the response to a Publish packets, depending on
// the fixed header it can be one of PUBACK, PUBREC, PUBREL or PUBCOMP
type PubAck struct {
	fixed bits

	packetID   wuint16
	reasonCode wuint8
	reason     wstring
	UserProperties
}

func (p *PubAck) String() string {
	return fmt.Sprintf("%s p%v %s%s %v bytes",
		firstByte(p.fixed).String(),
		p.packetID,
		ReasonCode(p.reasonCode).String(),
		func() string {
			if p.reasonCode > 0 && len(p.reason) > 0 {
				return " " + string(p.reason)
			}
			return ""
		}(),
		p.width(),
	)
}

func (p *PubAck) SetPacketID(v uint16) { p.packetID = wuint16(v) }
func (p *PubAck) PacketID() uint16     { return uint16(p.packetID) }

func (p *PubAck) SetReasonCode(v ReasonCode) { p.reasonCode = wuint8(v) }
func (p *PubAck) ReasonCode() ReasonCode     { return ReasonCode(p.reasonCode) }

func (p *PubAck) SetReason(v string) { p.reason = wstring(v) }
func (p *PubAck) Reason() string     { return string(p.reason) }

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
	i += p.reason.fillProp(b, i, ReasonString)
	i += p.UserProperties.properties(b, i)
	return i - n
}

func (p *PubAck) UnmarshalBinary(data []byte) error {
	b := &buffer{data: data}
	b.get(&p.packetID)
	// no more data, see 3.4.2.1 PUBACK Reason Code
	if len(data) > 2 {
		b.get(&p.reasonCode)
		b.getAny(p.propertyMap(), p.appendUserProperty)
	}
	return b.err
}

func (p *PubAck) propertyMap() map[Ident]wireType {
	return map[Ident]wireType{
		ReasonString: &p.reason,
	}
}
