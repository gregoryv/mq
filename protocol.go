package mq

import (
	"context"
	"encoding"
	"fmt"
	"io"
)

/*
Client implementations are responsible for

  - Sync writes and reads of packets
  - Add packet ID's and acknowledge packets
*/
type Client interface {
	Send(context.Context, Packet) error
}

// Handlers are used for both incoming and outgoing processing of
// packets.
type Handler func(context.Context, Packet) error

// PubHandler is specific to publish packets
type PubHandler func(context.Context, *Publish) error

type Middleware func(next Handler) Handler

// Packet and ControlPacket can be used interchangebly.
type Packet = ControlPacket

type ControlPacket interface {
	io.WriterTo
	encoding.BinaryUnmarshaler
	fmt.Stringer
}
