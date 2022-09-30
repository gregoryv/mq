package mq

import (
	"context"
)

// wip design the client and router

/*
Client implementations are responsible for

  - Sync writes and reads of packets
  - Add packet ID's and acknowledge packets
*/
type Client interface {
	Router
	// should they block until acked? if ack is expected
	Pub(context.Context, *Publish) error

	// Sub sends subscribe packets for all subscriptions in the router
	// that have not yet been send.
	Sub(context.Context) error
}

type Router interface {
	Add(Subscription)
}

type Subscription struct {
	*Subscribe
	Receiver
}

// Handler acts on incoming packets. Initially designed for the client
// side though could be used on the server aswell. Time will tell.
type Receiver func(Packet) error

// Packet represents any packet that can or should be handled by the
// application layer. Using a combined type for acknowledgements and
// publish control packets will hopefully make it easier to write
// handlers (todo remove this sentence) when done.
type Packet interface {
	Client() Client
	IsAck() bool

	// valid for Publish packets, ie. !IsAck()
	ContentType() string
	CorrelationData() []byte
	Duplicate() bool
	Payload() []byte
	PacketID() uint16
	ResponseTopic() string
}
