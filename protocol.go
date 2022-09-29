package mq

import (
	"context"
)

// wip design the client and router

/*
Client implementations are responsible for

1. Sync writes and reads of packets
2. Add packet ID's and acknowledge packets
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

type Handler interface {
	Act(context.Context, Packet) error
}

type HandlerFunc func(context.Context, Packet) error

func (h HandlerFunc) Act(ctx context.Context, p Packet) error {
	return h(ctx, p)
}

// Packet represents any packet that can or should be handled by the
// application layer.
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
