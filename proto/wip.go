package proto

import (
	"context"

	"github.com/gregoryv/mqtt"
)

// wip design the client and router

/*
Client implementations are responsible for

1. Sync writes and reads of packets
2. Add packet ID's and acknowledge packets
*/
type Client interface {
	// should they block until acked? if ack is expected
	Pub(context.Context, *mqtt.Publish) error
	Sub(context.Context, *mqtt.Subscribe, HandlerFunc) error
}

type Router interface {
	Add(Subscription)
}

type Subscription interface {
	Packet() *mqtt.Subscribe
	Handler
}

type Handler interface {
	Act(context.Context, Packet) error
}

type HandlerFunc func(context.Context, Packet)

func (h HandlerFunc) Act(ctx context.Context, p Packet) {
	h(ctx, p)
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
