package mq

import (
	"encoding"
	"fmt"
	"io"
)

func ReadPacket(r io.Reader) (ControlPacket, error) {
	var fh fixedHeader
	if _, err := fh.ReadFrom(r); err != nil {
		return nil, fmt.Errorf("ReadPacket: %w", err)
	}

	return fh.ReadRemaining(r)
}

// Packet and ControlPacket can be used interchangebly.
type Packet = ControlPacket

type ControlPacket interface {
	io.WriterTo
	encoding.BinaryUnmarshaler
	fmt.Stringer
}

type HasPacketID interface {
	PacketID() uint16
}

type fixedHeader struct {
	fixed        bits
	remainingLen vbint
}

// ReadFrom reads the fixed byte and the remaining length, use
// ReadRemaining for the rest.
//
// Note: Reason for splitting this up is, pahos Unpack works on the
// remaining only. Also it gives us possible ways of optimizing memory
// usage when reading packets, i.e. using shared FixedHeaders.
func (f *fixedHeader) ReadFrom(r io.Reader) (int64, error) {
	n, err := f.fixed.ReadFrom(r)
	if err != nil {
		return n, err
	}
	m, err := f.remainingLen.ReadFrom(r)
	return n + m, err
}

// ReadRemaining reads the reamining data and converts to a control
// packet.
func (f *fixedHeader) ReadRemaining(r io.Reader) (ControlPacket, error) {
	var p ControlPacket
	switch byte(f.fixed) & 0b1111_0000 {

	case PUBLISH:
		p = &Publish{fixed: f.fixed}

	case PUBREL:
		p = &PubRel{PubAck{fixed: f.fixed}}

	case PUBCOMP:
		p = &PubComp{fixed: f.fixed}

	case PUBREC:
		p = &PubRec{fixed: f.fixed}

	case PUBACK:
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
		return nil, fmt.Errorf(
			"%s ReadRemaining: %w",
			firstByte(f.fixed).String(), err,
		)
	}

	if err := p.UnmarshalBinary(data); err != nil {
		return nil, fmt.Errorf(
			"%s %v UnmarshalBinary: %w",
			firstByte(f.fixed).String(), f.remainingLen, err,
		)
	}
	return p, nil
}
