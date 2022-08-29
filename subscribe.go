package mqtt

import "fmt"

func NewSubscribe() Subscribe {
	return Subscribe{fixed: Bits(SUBSCRIBE)}
}

type Subscribe struct {
	fixed          Bits
	packetID       wuint16
	subscriptionID vbint
	userProp       []property
	filters        []TopicFilter
}

func (p *Subscribe) String() string {
	return fmt.Sprintf("%s %v",
		firstByte(p.fixed).String(),
		p.packetID,
	)
}

func (p *Subscribe) SetPacketID(v uint16) { p.packetID = wuint16(v) }
func (p *Subscribe) PacketID() uint16     { return uint16(p.packetID) }

func (p *Subscribe) SetSubscriptionID(v int) { p.subscriptionID = vbint(v) }
func (p *Subscribe) SubscriptionID() int     { return int(p.subscriptionID) }

// ----------------------------------------

type TopicFilter struct {
	filter  wstring
	options Bits
}
