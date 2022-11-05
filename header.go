package mq

import (
	"fmt"
	"io"
)

type fixedHeader struct {
	fixed        bits
	remainingLen vbint
}

// ReadFrom reads the fixed byte and the remaining length, use
// ReadRemaining for the rest.
//
// Note: Reason for splitting this up is that pahos Unpack works on
// the remaining only. Also it gives us possible ways of optimizing
// memory usage when reading packets, i.e. using shared FixedHeaders.
func (f *fixedHeader) ReadFrom(r io.Reader) (int64, error) {
	n, err := f.fixed.ReadFrom(r)
	if err != nil {
		return n, fmt.Errorf("ReadFrom: %w", err)
	}
	m, err := f.remainingLen.ReadFrom(r)
	if err != nil {
		return n + m, fmt.Errorf("ReadFrom: %w", err)
	}
	return n + m, nil
}

// ReadRemaining is more related to client and server
func (f *fixedHeader) ReadRemaining(r io.Reader) (ControlPacket, error) {
	var p ControlPacket
	switch byte(f.fixed) & 0b1111_0000 {

	case PUBLISH:
		p = &Publish{fixed: f.fixed}

	case PUBREL:
		p = &PubRel{PubAck{fixed: f.fixed}}

	case PUBACK, PUBREC, PUBCOMP:
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

	case AUTH:
		p = &Auth{fixed: f.fixed}

	default:
		p = &Undefined{}
	}
	if f.remainingLen == 0 {
		return p, nil
	}
	data := make([]byte, int(f.remainingLen))
	if _, err := r.Read(data); err != nil {
		return nil, fmt.Errorf("%s read remaining: %w", firstByte(f.fixed).String(), err)
	}

	if err := p.UnmarshalBinary(data); err != nil {
		return nil, fmt.Errorf("%s %v UnmarshalBinary: %w", firstByte(f.fixed).String(), f.remainingLen, err)
	}
	return p, nil
}
