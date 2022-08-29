package mqtt

import (
	"bytes"
	"fmt"
)

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
	return fmt.Sprintf("%s %v, %s",
		firstByte(p.fixed).String(),
		p.packetID,
		p.filters[0],
	)
}

func (p *Subscribe) SetPacketID(v uint16) { p.packetID = wuint16(v) }
func (p *Subscribe) PacketID() uint16     { return uint16(p.packetID) }

func (p *Subscribe) SetSubscriptionID(v int) { p.subscriptionID = vbint(v) }
func (p *Subscribe) SubscriptionID() int     { return int(p.subscriptionID) }

func (p *Subscribe) AddUserProp(key, val string) {
	p.AddUserProperty(property{key, val})
}
func (p *Subscribe) AddUserProperty(prop property) {
	p.userProp = append(p.userProp, prop)
}

func (p *Subscribe) AddFilter(filter string, options Fop) {
	p.filters = append(p.filters, TopicFilter{
		filter:  wstring(filter),
		options: Bits(options),
	})
}

// ----------------------------------------

type TopicFilter struct {
	filter  wstring
	options Bits
}

func (c TopicFilter) String() string {
	flags := bytes.Repeat([]byte("-"), 8)

	mark := func(i int, flag byte, v byte) {
		if !c.options.Has(flag) {
			return
		}
		flags[i] = v
	}

	// QoS
	mark(7, byte(FopQoS1), '1')
	mark(6, byte(FopQoS2), '2')
	if c.options.Has(byte(FopQoS3)) {
		flags[7] = '!'
		flags[6] = '!'
	}
	if c.options.Has(byte(FopNL)) {
		flags[5] = 'n'
	}
	if c.options.Has(byte(FopRAP)) {
		flags[4] = 'p'
	}
	// Retain
	flags[3] = '0'
	flags[2] = 'r'
	if c.options.Has(byte(FopRetain1)) {
		flags[3] = '1'
		flags[2] = 'r'
	}
	if c.options.Has(byte(FopRetain2)) {
		flags[3] = '2'
		flags[2] = 'r'
	}
	if c.options.Has(byte(FopRetain3)) {
		flags[3] = '!'
		flags[2] = '!'
	}

	// Reserved
	mark(1, 1<<6, '!')
	mark(0, 1<<7, '!')

	return fmt.Sprintf("%s %s", c.filter, string(flags))
}
