package tt

import (
	"github.com/gregoryv/mq"
)

func NewSubscription() *Subscription {
	return &Subscription{}
}

type Subscription struct {
	packet  *mq.Subscribe
	handler mq.Handler
}

func (s *Subscription) SetPacket(v *mq.Subscribe) { s.packet = v }
func (s *Subscription) Packet() *mq.Subscribe     { return s.packet }

func (s *Subscription) SetHandler(v mq.HandlerFunc) { s.handler = v }
func (s *Subscription) Handler() mq.Handler         { return s.handler }
