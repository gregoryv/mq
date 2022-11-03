package mq

import (
	"fmt"
	"io"
)

func NewPingReq() *PingReq {
	return &PingReq{fixed: bits(PINGREQ)}
}

type PingReq struct {
	fixed bits
}

func (p *PingReq) String() string {
	return fmt.Sprintf("%s %v bytes",
		firstByte(p.fixed).String(),
		p.width(),
	)
}

func (p *PingReq) WriteTo(w io.Writer) (int64, error) {
	b := make([]byte, p.width())
	p.fill(b, 0)
	n, err := w.Write(b)
	return int64(n), err
}

func (p *PingReq) width() int {
	return p.fill(_LEN, 0)
}

func (p *PingReq) fill(b []byte, i int) int {
	i += p.fixed.fill(b, i)  // firstByte header
	i += vbint(0).fill(b, i) // remaining length none
	return i
}

func (p *PingReq) UnmarshalBinary(data []byte) error {
	// there should not be any data
	return nil
}
