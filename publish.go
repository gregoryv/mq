package mqtt

import "fmt"

func NewPublish() *Publish {
	return &Publish{}
}

type Publish struct {
	fixed Bits
}

func (p *Publish) UnmarshalBinary(data []byte) error {
	return fmt.Errorf(": todo")
}

func (p *Publish) fill(b []byte, i int) int {
	panic("implement Publish.fill")
	return -1
}

func (p *Publish) width() int {
	return p.fill(_LEN, 0)
}
