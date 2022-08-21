package mqtt

import (
	"fmt"
	"io"
)

func NewPublish() *Publish {
	return &Publish{
		fixed: Bits(PUBLISH),
	}
}

type Publish struct {
	fixed     Bits
	topicName wstring

	packetIdent           wuint16
	topicAlias            wuint16
	messageExpiryInterval wuint32
	responseTopic         wstring
	correlationData       bindata
	userProp              []property
	subIdentifiers        []vbint
	contentType           wstring

	payloadFormat wbool
	payload       bindata
}

func (p *Publish) UnmarshalBinary(data []byte) error {
	return fmt.Errorf(": todo")
}

func (p *Publish) WriteTo(w io.Writer) (int64, error) {
	// allocate full size of entire packet
	b := make([]byte, p.fill(_LEN, 0))
	p.fill(b, 0)

	n, err := w.Write(b)
	return int64(n), err
}

func (p *Publish) fill(b []byte, i int) int {
	remainingLen := vbint(
		p.variableHeader(_LEN, 0) + p.payload.fill(_LEN, 0),
	)

	i += p.fixed.fill(b, i)      // firstByte header
	i += remainingLen.fill(b, i) // remaining length
	i += p.variableHeader(b, i)  // variable header
	i += p.payload.fill(b, i)    // payload

	return i
}
func (p *Publish) variableHeader(b []byte, i int) int {
	n := i
	// todo
	return i - n
}

func (p *Publish) width() int {
	return p.fill(_LEN, 0)
}

func (p *Publish) String() string {
	return fmt.Sprintf("%s", firstByte(p.fixed).String())
}
