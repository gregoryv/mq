package mqtt

import (
	"encoding"
	"fmt"
	"io"
)

func ReadPacket(r io.Reader) (ControlPacket, error) {
	var fh FixedHeader
	if _, err := fh.ReadFrom(r); err != nil {
		return nil, err
	}

	got, err := fh.ReadPacket(r)
	if err != nil {
		return nil, err
	}
	return got, nil
}

type FixedHeader struct {
	fixed        Bits
	remainingLen vbint
}

func (f *FixedHeader) ReadFrom(r io.Reader) (int64, error) {
	n, err := f.fixed.ReadFrom(r)
	if err != nil {
		return n, err
	}
	m, err := f.remainingLen.ReadFrom(r)
	return n + m, err
}

// ReadPacket is more related to client and server
func (f *FixedHeader) ReadPacket(r io.Reader) (ControlPacket, error) {
	data := make([]byte, int(f.remainingLen))
	if _, err := r.Read(data); err != nil {
		return nil, err
	}

	var p ControlPacket
	switch {

	case f.fixed.Has(PUBLISH):
		p = &Publish{fixed: f.fixed}

	case f.fixed.Has(CONNECT):
		p = &Connect{fixed: f.fixed}

	case f.fixed.Has(CONNACK):
		p = &ConnAck{fixed: f.fixed}

	default:
		panic(fmt.Sprintf("unknown %v", f.fixed))
	}

	if err := p.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return p, nil
}

type ControlPacket interface {
	io.WriterTo
	encoding.BinaryUnmarshaler
	fmt.Stringer
}
