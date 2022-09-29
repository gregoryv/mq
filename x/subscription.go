package x

import (
	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/proto"
)

func NewSubscription() *Subscription {
	return &Subscription{}
}

type Subscription struct {
	packet  *mq.Subscribe
	handler proto.Handler
}

func (s *Subscription) SetPacket(v *mq.Subscribe) { s.packet = v }
func (s *Subscription) Packet() *mq.Subscribe     { return s.packet }

func (s *Subscription) SetHandler(v proto.HandlerFunc) { s.handler = v }
func (s *Subscription) Handler() proto.Handler         { return s.handler }
