package mq

import (
	"fmt"
	"io"
)

func NewUndefined() Undefined {
	return Undefined{}
}

type Undefined struct {
	fixed Bits
}

func (p *Undefined) String() string {
	return fmt.Sprintf("%s %v bytes",
		firstByte(p.fixed).String(), 0,
	)
}

func (p *Undefined) WriteTo(w io.Writer) (int64, error) {
	return 0, fmt.Errorf("Undefined cannot be written")
}

func (p *Undefined) UnmarshalBinary(data []byte) error {
	// there should not be any data
	return nil
}
