package mqtt

import (
	"fmt"
	"io"
)

func NewPublish() *Publish {
	return &Publish{
		fixed: Bits(PUBLISH),
	}
}

type Publish struct {
	fixed     Bits
	topicName wstring

	packetID wuint16 // packet identifier

	topicAlias            wuint16
	messageExpiryInterval wuint32
	responseTopic         wstring
	correlationData       bindata
	userProp              []property
	subscriptionIDs       []uint32
	contentType           wstring

	payloadFormat wbool
	payload       bindata
}

func (p *Publish) SetDuplicate(v bool) { p.fixed.toggle(DUP, v) }
func (p *Publish) Duplicate() bool     { return p.fixed.Has(DUP) }

func (p *Publish) SetRetain(v bool) { p.fixed.toggle(RETAIN, v) }
func (p *Publish) Retain() bool     { return p.fixed.Has(RETAIN) }

// SetQoS, 1 or 2 other values unset the QoS
func (p *Publish) SetQoS(v uint8) {
	p.fixed &= Bits(^(QoS3)) // reset
	switch v {
	case 1:
		p.fixed.toggle(QoS1, true)
	case 2:
		p.fixed.toggle(QoS2, true)
	}
}

func (p *Publish) QoS() uint8 {
	switch {
	case p.fixed.Has(QoS3):
		return 3 // malformed
	case p.fixed.Has(QoS1):
		return 1
	case p.fixed.Has(QoS2):
		return 2
	}
	return 0
}

func (p *Publish) SetTopicName(v string) { p.topicName = wstring(v) }
func (p *Publish) TopicName() string     { return string(p.topicName) }

func (p *Publish) SetPacketID(v uint16) { p.packetID = wuint16(v) }
func (p *Publish) PacketID() uint16     { return uint16(p.packetID) }

func (p *Publish) SetPayloadFormat(v bool) { p.payloadFormat = wbool(v) }
func (p *Publish) PayloadFormat() bool     { return bool(p.payloadFormat) }

func (p *Publish) SetMessageExpiryInterval(v uint32) {
	p.messageExpiryInterval = wuint32(v)
}
func (p *Publish) MessageExpiryInterval() uint32 {
	return uint32(p.messageExpiryInterval)
}

func (p *Publish) SetTopicAlias(v uint16) { p.topicAlias = wuint16(v) }
func (p *Publish) TopicAlias() uint16     { return uint16(p.topicAlias) }

func (p *Publish) SetResponseTopic(v string) { p.responseTopic = wstring(v) }
func (p *Publish) ResponseTopic() string     { return string(p.responseTopic) }

func (p *Publish) SetCorrelationData(v []byte) { p.correlationData = bindata(v) }
func (p *Publish) CorrelationData() []byte     { return []byte(p.correlationData) }

// AddUserProp adds a user property. The User Property is allowed to
// appear multiple times to represent multiple name, value pairs. The
// same name is allowed to appear more than once.
func (p *Publish) AddUserProp(key, val string) {
	p.AddUserProperty(property{key, val})
}
func (p *Publish) AddUserProperty(prop property) {
	p.appendUserProperty(prop)
}
func (p *Publish) appendUserProperty(prop property) {
	p.userProp = append(p.userProp, prop)
}

func (p *Publish) AddSubscriptionID(v uint32) {
	p.subscriptionIDs = append(p.subscriptionIDs, v)
}

func (p *Publish) SubscriptionIDs() []uint32 {
	return p.subscriptionIDs
}

func (p *Publish) SetContentType(v string) { p.contentType = wstring(v) }
func (p *Publish) ContentType() string     { return string(p.contentType) }

func (p *Publish) SetPayload(v []byte) { p.payload = bindata(v) }
func (p *Publish) Payload() []byte     { return []byte(p.payload) }

// end settings
// ----------------------------------------

func (p *Publish) UnmarshalBinary(data []byte) error {
	return fmt.Errorf(": todo")
}

func (p *Publish) WriteTo(w io.Writer) (int64, error) {
	// allocate full size of entire packet
	b := make([]byte, p.fill(_LEN, 0))
	p.fill(b, 0)

	n, err := w.Write(b)
	return int64(n), err
}

func (p *Publish) fill(b []byte, i int) int {
	remainingLen := vbint(p.variableHeader(_LEN, 0))

	if len(p.payload) > 0 {
		remainingLen += vbint(p.payload.fill(_LEN, 0))
	}

	i += p.fixed.fill(b, i)      // FirstByte header
	i += remainingLen.fill(b, i) // remaining length
	i += p.variableHeader(b, i)  // variable header
	if len(p.payload) > 0 {
		i += p.payload.fill(b, i) // payload
	}

	return i
}
func (p *Publish) variableHeader(b []byte, i int) int {
	n := i

	i += p.topicName.fill(b, i)
	if v := p.QoS(); v == 1 || v == 2 {
		i += p.packetID.fill(b, i)
	}
	i += vbint(p.properties(_LEN, 0)).fill(b, i) // Properties len
	i += p.properties(b, i)                      // Properties

	return i - n
}

func (p *Publish) properties(b []byte, i int) int {
	n := i
	fill := func(id Ident, v wireType) {
		i += id.fill(b, i)
		i += v.fill(b, i)
	}

	if v := p.payloadFormat; v {
		fill(PayloadFormatIndicator, &v)
	}
	if v := p.messageExpiryInterval; v > 0 {
		fill(MessageExpiryInterval, &v)
	}
	if v := p.topicAlias; v > 0 {
		fill(TopicAlias, &v)
	}
	if v := p.responseTopic; len(v) > 0 {
		fill(ResponseTopic, &v)
	}
	if v := p.correlationData; len(v) > 0 {
		fill(CorrelationData, &v)
	}

	for _, v := range p.userProp {
		fill(UserProperty, &v)
	}

	for _, v := range p.subscriptionIDs {
		if v > 0 {
			id := vbint(v)
			fill(SubscriptionID, &id)
		}
	}
	if v := p.contentType; len(v) > 0 {
		fill(ContentType, &v)
	}
	return i - n
}

func (p *Publish) width() int {
	return p.fill(_LEN, 0)
}

func (p *Publish) String() string {
	return fmt.Sprintf("%s %v bytes",
		FirstByte(p.fixed).String(),
		p.width(),
	)
}
