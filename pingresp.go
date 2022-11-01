package mq

import (
	"fmt"
	"io"
)

func NewPingResp() *PingResp {
	return &PingResp{fixed: Bits(PINGRESP)}
}

type PingResp struct {
	fixed Bits
}

func (p *PingResp) String() string {
	return fmt.Sprintf("%s %v bytes",
		firstByte(p.fixed).String(),
		p.width(),
	)
}

func (p *PingResp) WriteTo(w io.Writer) (int64, error) {
	b := make([]byte, p.width())
	p.fill(b, 0)
	n, err := w.Write(b)
	return int64(n), err
}

func (p *PingResp) width() int {
	return p.fill(_LEN, 0)
}

func (p *PingResp) fill(b []byte, i int) int {
	i += p.fixed.fill(b, i)  // firstByte header
	i += vbint(0).fill(b, i) // remaining length none
	return i
}

func (p *PingResp) UnmarshalBinary(data []byte) error {
	// there should not be any data
	return nil
}
