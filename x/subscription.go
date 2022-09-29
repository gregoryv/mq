package x

import (
	"github.com/gregoryv/mqtt"
	"github.com/gregoryv/mqtt/proto"
)

func NewSubscription() *Subscription {
	return &Subscription{}
}

type Subscription struct {
	packet  *mqtt.Subscribe
	handler proto.Handler
}

func (s *Subscription) SetPacket(v *mqtt.Subscribe) { s.packet = v }
func (s *Subscription) Packet() *mqtt.Subscribe     { return s.packet }

func (s *Subscription) SetHandler(v proto.HandlerFunc) { s.handler = v }
func (s *Subscription) Handler() proto.Handler         { return s.handler }
