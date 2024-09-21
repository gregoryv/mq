package mq

import (
	"fmt"
	"io"
)

// NewPubRec returns control packet with type PUBREC
func NewPubRec() *PubRec {
	return &PubRec{fixed: bits(PUBREC)}
}

// A PubRec packet is the response to a Publish packets, depending on
// the fixed header it can be one of PUBACK, PUBREC, PUBREL or PUBCOMP
type PubRec struct {
	fixed bits

	packetID   wuint16
	reasonCode wuint8
	reason     wstring
	UserProperties
}

func (p *PubRec) String() string {
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

func (p *PubRec) dump(w io.Writer) {
	fmt.Fprintf(w, "PacketID: %v\n", p.PacketID())
	fmt.Fprintf(w, "Reason: %v\n", p.ReasonString())
	fmt.Fprintf(w, "ReasonCode: %v\n", p.ReasonCode())
	p.UserProperties.dump(w)
}

func (p *PubRec) SetPacketID(v uint16) { p.packetID = wuint16(v) }
func (p *PubRec) PacketID() uint16     { return uint16(p.packetID) }

func (p *PubRec) SetReasonCode(v ReasonCode) { p.reasonCode = wuint8(v) }
func (p *PubRec) ReasonCode() ReasonCode     { return ReasonCode(p.reasonCode) }

func (p *PubRec) SetReasonString(v string) { p.reason = wstring(v) }
func (p *PubRec) ReasonString() string     { return string(p.reason) }

func (p *PubRec) WriteTo(w io.Writer) (int64, error) {
	b := make([]byte, p.fill(_LEN, 0))
	p.fill(b, 0)
	n, err := w.Write(b)
	return int64(n), err
}

func (p *PubRec) width() int {
	return p.fill(_LEN, 0)
}

func (p *PubRec) fill(b []byte, i int) int {
	remainingLen := vbint(p.variableHeader(_LEN, 0))
	i += p.fixed.fill(b, i)      // firstByte header
	i += remainingLen.fill(b, i) // remaining length
	i += p.variableHeader(b, i)
	return i
}

func (p *PubRec) variableHeader(b []byte, i int) int {
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

func (p *PubRec) properties(b []byte, i int) int {
	n := i
	i += p.reason.fillProp(b, i, ReasonString)
	i += p.UserProperties.properties(b, i)
	return i - n
}

func (p *PubRec) UnmarshalBinary(data []byte) error {
	b := &buffer{data: data}
	b.get(&p.packetID)
	// no more data, see 3.4.2.1 PUBACK Reason Code
	if len(data) > 2 {
		b.get(&p.reasonCode)
		b.getAny(p.propertyMap(), p.appendUserProperty)
	}
	return b.err
}

func (p *PubRec) propertyMap() map[Ident]func() wireType {
	return map[Ident]func() wireType{
		ReasonString: func() wireType { return &p.reason },
	}
}
