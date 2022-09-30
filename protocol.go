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
	Connect(context.Context, *Connect) error
	Disconnect(context.Context, *Disconnect) error

	// Pub writes the given control packet on the wire, fails if could
	// not be written. The call does not wait for a PubAck, see
	// Receiver.
	Pub(context.Context, *Publish) error

	// Sub writes the given control packet on the wire, fails if could
	// not be written. The call does not wait for a SubAck, see
	// Receiver.
	Sub(context.Context, *Subscribe) error
}

type Router interface {
	Add(...Subscription)
	Subscriptions() []*Subscription
}

type Subscription struct {
	*Subscribe
	Receiver
}

// Receiver is called on incoming packets. Initially designed for the
// client side.
type Receiver func(Packet) error

// Packet and ControlPacket can be used interchangebly.
type Packet = ControlPacket

type ControlPacket interface {
	io.WriterTo
	encoding.BinaryUnmarshaler
	fmt.Stringer
}
