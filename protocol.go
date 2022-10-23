package mq

import (
	"context"
	"encoding"
	"fmt"
	"io"
)

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

// Packet and ControlPacket can be used interchangebly.
type Packet = ControlPacket

type ControlPacket interface {
	io.WriterTo
	encoding.BinaryUnmarshaler
	fmt.Stringer
}

// Handlers are used for both incoming and outgoing processing of
// packets.
type Handler func(context.Context, Packet) error

// PubHandler is specific to publish packets
type PubHandler func(context.Context, *Publish) error

type Middleware func(next Handler) Handler

type HasPacketID interface {
	PacketID() uint16
}
