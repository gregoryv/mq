package proto

import (
	"context"

	"github.com/gregoryv/mqtt"
)

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
	Act(Client, Packet)
}

type HandlerFunc func(Client, Packet)

func (h HandlerFunc) Act(c Client, p Packet) {
	h(c, p)
}

// Packet represents any packet that can or should be handled by the
// application layer.
type Packet interface {
	IsAck() bool

	// valid for Publish packets, ie. !IsAck()
	ContentType() string
	CorrelationData() []byte
	Duplicate() bool
	Payload() []byte
	PacketID() uint16
	ResponseTopic() string
}
