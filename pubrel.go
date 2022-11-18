package mq

import (
	"fmt"
	"io"
)

// NewPubRel returns control packet with type PUBREL
func NewPubRel() *PubRel {
	p := &PubRel{}
	p.fixed = bits(PUBREL | 1<<1)
	return p
}

// A PubRel packet is the response to a Publish packets, depending on
// the fixed header it can be one of PUBACK, PUBREC, PUBREL or PUBCOMP
type PubRel struct {
	fixed bits

	packetID   wuint16
	reasonCode wuint8
	reason     wstring
	UserProperties
}

func (p *PubRel) String() string {
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

func (p *PubRel) SetPacketID(v uint16) { p.packetID = wuint16(v) }
func (p *PubRel) PacketID() uint16     { return uint16(p.packetID) }

func (p *PubRel) SetReasonCode(v ReasonCode) { p.reasonCode = wuint8(v) }
func (p *PubRel) ReasonCode() ReasonCode     { return ReasonCode(p.reasonCode) }

func (p *PubRel) SetReason(v string) { p.reason = wstring(v) }
func (p *PubRel) Reason() string     { return string(p.reason) }

func (p *PubRel) WriteTo(w io.Writer) (int64, error) {
	b := make([]byte, p.fill(_LEN, 0))
	p.fill(b, 0)
	n, err := w.Write(b)
	return int64(n), err
}

func (p *PubRel) width() int {
	return p.fill(_LEN, 0)
}

func (p *PubRel) fill(b []byte, i int) int {
	remainingLen := vbint(p.variableHeader(_LEN, 0))
	i += p.fixed.fill(b, i)      // firstByte header
	i += remainingLen.fill(b, i) // remaining length
	i += p.variableHeader(b, i)
	return i
}

func (p *PubRel) variableHeader(b []byte, i int) int {
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

func (p *PubRel) properties(b []byte, i int) int {
	n := i
	i += p.reason.fillProp(b, i, ReasonString)
	i += p.UserProperties.properties(b, i)
	return i - n
}

func (p *PubRel) UnmarshalBinary(data []byte) error {
	b := &buffer{data: data}
	b.get(&p.packetID)
	// no more data, see 3.4.2.1 PUBACK Reason Code
	if len(data) > 2 {
		b.get(&p.reasonCode)
		b.getAny(p.propertyMap(), p.appendUserProperty)
	}
	return b.err
}

func (p *PubRel) propertyMap() map[Ident]wireType {
	return map[Ident]wireType{
		ReasonString: &p.reason,
	}
}
