package mqtt

import (
	"encoding"
	"fmt"
	"io"
)

type ControlPacket interface {
	io.WriterTo
	encoding.BinaryUnmarshaler
	fmt.Stringer
}

func ReadPacket(r io.Reader) (ControlPacket, error) {
	var fh FixedHeader
	if _, err := fh.ReadFrom(r); err != nil {
		return nil, err
	}

	got, err := fh.ReadRemaining(r)
	if err != nil {
		return nil, err
	}
	return got, nil
}

type FixedHeader struct {
	fixed        Bits
	remainingLen vbint
}

// ReadFrom reads the fixed byte and the remaining length, use
// ReadRemaining for the rest.
//
// Note: Reason for splitting this up is that pahos Unpack works on
// the remaining only. Also it gives us possible ways of optimizing
// memory usage when reading packets, i.e. using shared FixedHeaders.
func (f *FixedHeader) ReadFrom(r io.Reader) (int64, error) {
	n, err := f.fixed.ReadFrom(r)
	if err != nil {
		return n, err
	}
	m, err := f.remainingLen.ReadFrom(r)
	return n + m, err
}

// ReadRemaining is more related to client and server
func (f *FixedHeader) ReadRemaining(r io.Reader) (ControlPacket, error) {
	data := make([]byte, int(f.remainingLen))
	if _, err := r.Read(data); err != nil {
		return nil, err
	}

	var p ControlPacket
	switch byte(f.fixed) & 0b1111_0000 {

	case PUBLISH:
		p = &Publish{fixed: f.fixed}

	case PUBACK, PUBREC, PUBREL, PUBCOMP:
		p = &PubAck{fixed: f.fixed}

	case CONNECT:
		p = &Connect{fixed: f.fixed}

	case CONNACK:
		p = &ConnAck{fixed: f.fixed}

	case SUBSCRIBE:
		p = &Subscribe{fixed: f.fixed}

	case UNSUBSCRIBE:
		p = &Unsubscribe{fixed: f.fixed}

	case SUBACK:
		p = &SubAck{fixed: f.fixed}

	case UNSUBACK:
		p = &UnsubAck{fixed: f.fixed}

	case PINGREQ:
		p = &PingReq{fixed: f.fixed}

	case PINGRESP:
		p = &PingResp{fixed: f.fixed}

	case DISCONNECT:
		p = &Disconnect{fixed: f.fixed}

	default:
		panic(fmt.Sprintf("ReadRemaining unhandled packet type %v", f.fixed))
	}

	if err := p.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return p, nil
}
