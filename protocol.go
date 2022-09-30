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
	// should they block until acked? if ack is expected
	Pub(context.Context, *Publish) error
	Sub(context.Context, *Subscribe, HandlerFunc) error
}

type Router interface {
	Add(Subscription)
}

type Subscription interface {
	Packet() *Subscribe
	Handler
}

// Handler acts on incoming packets. Initially designed for the client
// side though could be used on the server aswell. Time will tell.
type Handler interface {
	Act(Packet) error
}

type HandlerFunc func( Packet) error

func (h HandlerFunc) Act( p Packet) error {
	return h( p)
}

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
