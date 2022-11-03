package mq

import (
	"fmt"
	"io"
)

func NewUnsubscribe() *Unsubscribe {
	// wonder why bit 1 needs to be set? specification doesn't say
	return &Unsubscribe{fixed: bits(UNSUBSCRIBE | 1<<1)}
}

type Unsubscribe struct {
	fixed    bits
	packetID wuint16

	UserProperties
	filters []wstring
}

func (p *Unsubscribe) String() string {
	return fmt.Sprintf("%s p%v, %s, %v bytes",
		firstByte(p.fixed).String(),
		p.packetID,
		p.filterString(),
		p.width(),
	)
}

func (p *Unsubscribe) filterString() string {
	if len(p.filters) == 0 {
		return "no filters!" // malformed
	}
	return string(p.filters[0])
}

func (p *Unsubscribe) SetPacketID(v uint16) { p.packetID = wuint16(v) }
func (p *Unsubscribe) PacketID() uint16     { return uint16(p.packetID) }

func (p *Unsubscribe) AddFilter(filter string) {
	p.filters = append(p.filters, wstring(filter))
}

func (p *Unsubscribe) WriteTo(w io.Writer) (int64, error) {
	b := make([]byte, p.width())
	p.fill(b, 0)
	n, err := w.Write(b)
	return int64(n), err
}

func (p *Unsubscribe) width() int {
	return p.fill(_LEN, 0)
}

func (p *Unsubscribe) fill(b []byte, i int) int {
	remainingLen := vbint(
		p.variableHeader(_LEN, 0) + p.payload(_LEN, 0),
	)
	i += p.fixed.fill(b, i)      // firstByte header
	i += remainingLen.fill(b, i) // remaining length
	i += p.variableHeader(b, i)
	i += p.payload(b, i)

	return i
}

func (p *Unsubscribe) variableHeader(b []byte, i int) int {
	n := i
	i += p.packetID.fill(b, i)
	i += vbint(p.properties(_LEN, 0)).fill(b, i)
	i += p.properties(b, i)
	return i - n
}

func (p *Unsubscribe) payload(b []byte, i int) int {
	n := i
	for j, _ := range p.filters {
		i += p.filters[j].fill(b, i)
	}
	return i - n
}

func (p *Unsubscribe) UnmarshalBinary(data []byte) error {
	b := &buffer{data: data}
	b.get(&p.packetID)
	b.getAny(nil, p.appendUserProperty)

	for {
		var f wstring
		b.get(&f)
		p.filters = append(p.filters, f)
		if b.i == len(data) {
			break
		}
	}
	return b.err
}
