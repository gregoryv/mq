package mq

import (
	"context"
)

/*
Client implementations are responsible for

  - Sync writes and reads of packets
  - Add packet ID's and acknowledge packets
*/
type Client interface {
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

// Packet represents any packet that can or should be handled by the
// application layer. Using a combined type for acknowledgements and
// publish control packets will hopefully make it easier to write
// receivers (todo remove this sentence) when done.
type Packet interface {
	Client() Client

	// IsAck returns true if the packet is of ConnAck, PubAck, SubAck
	// or UnsubAck.
	IsAck() bool

	// valid for Publish packets, ie. !IsAck()
	ContentType() string
	CorrelationData() []byte
	Duplicate() bool
	Payload() []byte
	PacketID() uint16
	ResponseTopic() string
}
