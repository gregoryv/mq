package mq

import (
	"fmt"
	"io"
)

// Pub is a convenience method for creating a publish packet.
func Pub(qos uint8, topic, payload string) *Publish {
	p := NewPublish()
	p.SetQoS(qos)
	p.SetTopicName(topic)
	p.SetPayload([]byte(payload))
	return p
}

func NewPublish() *Publish {
	return &Publish{
		fixed: bits(PUBLISH),
	}
}

type Publish struct {
	// first fields are aligned for memory
	fixed         bits
	packetID      wuint16
	topicAlias    wuint16
	payloadFormat wbool

	messageExpiryInterval wuint32
	topicName             wstring
	responseTopic         wstring
	correlationData       bindata
	contentType           wstring
	payload               rawdata
	UserProperties
	subscriptionIDs []uint32
}

func (p *Publish) String() string {
	topic := string(p.topicName)
	if v := uint16(p.topicAlias); v > 0 {
		topic = fmt.Sprintf("topic:%v", v)
	}

	return withForm(p, fmt.Sprintf("%s p%v %s%s %v bytes",
		firstByte(p.fixed).String(),
		p.packetID,
		topic,
		func() string {
			if len(p.correlationData) == 0 {
				return ""
			}
			return " " + string(p.correlationData)
		}(),
		p.width(),
	))
}

func (p *Publish) dump(w io.Writer) {
	fmt.Fprintf(w, "ContentType: %v\n", p.ContentType())
	fmt.Fprintf(w, "CorrelationData: %v\n", p.CorrelationData())
	fmt.Fprintf(w, "Duplicate: %v\n", p.Duplicate())
	fmt.Fprintf(w, "MessageExpiryInterval: %v\n", p.MessageExpiryInterval())
	fmt.Fprintf(w, "PacketID: %v\n", p.PacketID())
	fmt.Fprintf(w, "Payload: %v\n", p.Payload())
	fmt.Fprintf(w, "PayloadFormat: %v\n", p.PayloadFormat())
	fmt.Fprintf(w, "QoS: %v\n", p.QoS())
	fmt.Fprintf(w, "ResponseTopic: %v\n", p.ResponseTopic())
	fmt.Fprintf(w, "Retain: %v\n", p.Retain())
	fmt.Fprintf(w, "SubscriptionIDs: %v\n", p.SubscriptionIDs())
	fmt.Fprintf(w, "TopicAlias: %v\n", p.TopicAlias())
	fmt.Fprintf(w, "TopicName: %v\n", p.TopicName())

	p.UserProperties.dump(w)
}

// WellFormed returns a Malformed error if the packet does not follow
// the specification.
func (p *Publish) WellFormed() *Malformed {
	if len(p.topicName) == 0 {
		return newMalformed(p, "topic name", "empty")
	}
	switch p.QoS() {
	case 1, 2:
		if p.packetID == 0 {
			return newMalformed(p, "packet ID", "empty")
		}
	case 3:
		return newMalformed(p, "QoS", "invalid")
	}

	return nil
}

func (p *Publish) SetDuplicate(v bool) { p.fixed.toggle(DUP, v) }
func (p *Publish) Duplicate() bool     { return p.fixed.Has(DUP) }

func (p *Publish) SetRetain(v bool) { p.fixed.toggle(RETAIN, v) }
func (p *Publish) Retain() bool     { return p.fixed.Has(RETAIN) }

// SetQoS, 0,1,2 or 3 other values unset the QoS. 3 is malformed but
// allowed to be set here.
func (p *Publish) SetQoS(v uint8) {
	p.fixed &= bits(^(QoS3)) // reset
	switch v {
	case 1:
		p.fixed.toggle(QoS1, true)
	case 2:
		p.fixed.toggle(QoS2, true)
	case 3:
		p.fixed.toggle(QoS3, true)
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

// SetPayloadFormat, false indicates that the message is unspecified
// bytes. True indicates that the message is UTF-8 encoded character
// data.
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

func (p *Publish) AddSubscriptionID(v uint32) {
	p.subscriptionIDs = append(p.subscriptionIDs, v)
}

func (p *Publish) SubscriptionIDs() []uint32 {
	return p.subscriptionIDs
}

// The value of the Content Type is defined by the sending and
// receiving application, e.g. it may be a mime type like
// application/json.
func (p *Publish) SetContentType(v string) { p.contentType = wstring(v) }
func (p *Publish) ContentType() string     { return string(p.contentType) }

func (p *Publish) SetPayload(v []byte) { p.payload = rawdata(v) }
func (p *Publish) Payload() []byte     { return []byte(p.payload) }

// end settings
// ----------------------------------------

func (p *Publish) WriteTo(w io.Writer) (int64, error) {
	b := make([]byte, p.fill(_LEN, 0))
	p.fill(b, 0)
	n, err := w.Write(b)
	return int64(n), err
}

func (p *Publish) width() int {
	return p.fill(_LEN, 0)
}

func (p *Publish) fill(b []byte, i int) int {
	remainingLen := vbint(p.variableHeader(_LEN, 0))

	if len(p.payload) > 0 {
		remainingLen += vbint(p.payload.fill(_LEN, 0))
	}

	i += p.fixed.fill(b, i)      // firstByte header
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
	i += p.payloadFormat.fillProp(b, i, PayloadFormatIndicator)
	i += p.messageExpiryInterval.fillProp(b, i, MessageExpiryInterval)
	i += p.topicAlias.fillProp(b, i, TopicAlias)
	i += p.responseTopic.fillProp(b, i, ResponseTopic)
	i += p.correlationData.fillProp(b, i, CorrelationData)
	i += p.contentType.fillProp(b, i, ContentType)

	i += p.UserProperties.properties(b, i)
	for j, _ := range p.subscriptionIDs {
		i += vbint(p.subscriptionIDs[j]).fillProp(b, i, SubscriptionID)
	}
	return i - n
}

func (p *Publish) UnmarshalBinary(data []byte) error {
	buf := &buffer{
		data:              data,
		addSubscriptionID: p.AddSubscriptionID,
	}
	get := buf.get

	get(&p.topicName)
	if v := p.QoS(); v == 1 || v == 2 {
		get(&p.packetID)
	}

	buf.getAny(p.propertyMap(), p.appendUserProperty)

	if len(data) > buf.i {
		get(&p.payload)
	}
	return buf.err
}

func (p *Publish) propertyMap() map[Ident]func() wireType {
	return map[Ident]func() wireType{
		PayloadFormatIndicator: func() wireType { return &p.payloadFormat },
		MessageExpiryInterval:  func() wireType { return &p.messageExpiryInterval },
		TopicAlias:             func() wireType { return &p.topicAlias },
		ResponseTopic:          func() wireType { return &p.responseTopic },
		CorrelationData:        func() wireType { return &p.correlationData },
		ContentType:            func() wireType { return &p.contentType },
	}
}
