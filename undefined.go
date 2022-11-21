package mq

import (
	"fmt"
	"io"
)

// Undefined represents a packet with type value of 0.
type Undefined struct {
	fixed bits
	data  []byte
}

func (p *Undefined) String() string {
	return fmt.Sprintf("%s %v bytes",
		firstByte(p.fixed).String(), 0,
	)
}

func (p *Undefined) Data() []byte { return p.data }

func (p *Undefined) WriteTo(w io.Writer) (int64, error) {
	return 0, fmt.Errorf("cannot write %T", p)
}

func (p *Undefined) UnmarshalBinary(data []byte) error {
	p.data = data
	return nil
}
