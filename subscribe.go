package mq

import (
	"bytes"
	"fmt"
	"io"
)

func NewSubscribe() *Subscribe {
	// wonder why bit 1 needs to be set? specification doesn't say
	return &Subscribe{fixed: bits(SUBSCRIBE | 1<<1)}
}

type Subscribe struct {
	fixed          bits
	packetID       wuint16
	subscriptionID vbint
	UserProperties
	filters []TopicFilter
}

func (p *Subscribe) String() string {
	return fmt.Sprintf("%s p%v, %s %v bytes",
		firstByte(p.fixed).String(),
		p.packetID,
		p.filterString(),
		p.width(),
	)
}

func (p *Subscribe) dump(w io.Writer) {
	fmt.Fprintf(w, "PacketID: %v\n", p.PacketID())
	fmt.Fprintf(w, "SubscriptionID: %v\n", p.SubscriptionID())

	if len(p.filters) > 0 {
		fmt.Fprintln(w, "Filters")
		for i, f := range p.filters {
			fmt.Fprintf(w, "  %v. %s\n", i, f)
		}
	}
	p.UserProperties.dump(w)
}

// filterString returns string representing filters for use in String
func (p *Subscribe) filterString() string {
	if len(p.filters) == 0 {
		return "no filters!" // malformed
	}
	return p.filters[0].String()
}

func (p *Subscribe) SetPacketID(v uint16) { p.packetID = wuint16(v) }
func (p *Subscribe) PacketID() uint16     { return uint16(p.packetID) }

func (p *Subscribe) SetSubscriptionID(v int) { p.subscriptionID = vbint(v) }
func (p *Subscribe) SubscriptionID() int     { return int(p.subscriptionID) }

func (p *Subscribe) AddFilter(filter string, options Opt) {
	p.filters = append(p.filters, TopicFilter{
		filter:  wstring(filter),
		options: bits(options),
	})
}
func (p *Subscribe) Filters() []TopicFilter {
	return p.filters
}

func (p *Subscribe) WriteTo(w io.Writer) (int64, error) {
	b := make([]byte, p.width())
	p.fill(b, 0)
	n, err := w.Write(b)
	return int64(n), err
}

func (p *Subscribe) width() int {
	return p.fill(_LEN, 0)
}

func (p *Subscribe) fill(b []byte, i int) int {
	remainingLen := vbint(
		p.variableHeader(_LEN, 0) + p.payload(_LEN, 0),
	)
	i += p.fixed.fill(b, i)      // firstByte header
	i += remainingLen.fill(b, i) // remaining length
	i += p.variableHeader(b, i)
	i += p.payload(b, i)

	return i
}

func (p *Subscribe) variableHeader(b []byte, i int) int {
	n := i
	i += p.packetID.fill(b, i)
	i += vbint(p.properties(_LEN, 0)).fill(b, i)
	i += p.properties(b, i)
	return i - n
}

func (p *Subscribe) properties(b []byte, i int) int {
	n := i
	for id, v := range p.propertyMap() {
		i += v.fillProp(b, i, id)
	}
	i += p.UserProperties.properties(b, i)
	return i - n
}

func (p *Subscribe) payload(b []byte, i int) int {
	n := i
	for j, _ := range p.filters {
		i += p.filters[j].fill(b, i)
	}
	return i - n
}

func (p *Subscribe) UnmarshalBinary(data []byte) error {
	b := &buffer{data: data}
	b.get(&p.packetID)
	b.getAny(p.propertyMap(), p.appendUserProperty)

	for {
		var f TopicFilter
		b.get(&f.filter)
		b.get(&f.options)
		p.filters = append(p.filters, f)
		if b.i == len(data) {
			break
		}
	}
	return b.err
}

func (p *Subscribe) propertyMap() map[Ident]wireType {
	return map[Ident]wireType{
		SubscriptionID: &p.subscriptionID,
	}
}

// ----------------------------------------

func NewTopicFilter(filter string, options Opt) TopicFilter {
	return TopicFilter{
		filter:  wstring(filter),
		options: bits(options),
	}
}

type TopicFilter struct {
	filter  wstring
	options bits
}

func (c TopicFilter) fill(b []byte, i int) int {
	n := i
	i += c.filter.fill(b, i)
	i += c.options.fill(b, i)
	return i - n
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
	mark(7, byte(OptQoS1), '1')
	mark(6, byte(OptQoS2), '2')
	if c.options.Has(byte(OptQoS3)) {
		flags[7] = '!'
		flags[6] = '!'
	}
	if c.options.Has(byte(OptNL)) {
		flags[5] = 'n'
	}
	if c.options.Has(byte(OptRAP)) {
		flags[4] = 'p'
	}
	// Retain
	flags[3] = '0'
	flags[2] = 'r'
	if c.options.Has(byte(OptRetain1)) {
		flags[3] = '1'
		flags[2] = 'r'
	}
	if c.options.Has(byte(OptRetain2)) {
		flags[3] = '2'
		flags[2] = 'r'
	}
	if c.options.Has(byte(OptRetain3)) {
		flags[3] = '!'
		flags[2] = '!'
	}

	// Reserved
	mark(1, 1<<6, '!')
	mark(0, 1<<7, '!')

	return fmt.Sprintf("%s %s", c.filter, string(flags))
}
