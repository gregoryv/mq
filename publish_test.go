package mqtt

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"testing"
)

func TestPublish(t *testing.T) {
	p := NewPublish()
	eq(t, p.SetRetain, p.Retain, true)
	eq(t, p.SetQoS, p.QoS, 1)
	eq(t, p.SetDuplicate, p.Duplicate, true)

	eq(t, p.SetTopicName, p.TopicName, "topic/temp")
	eq(t, p.SetPacketID, p.PacketID, 1)
	eq(t, p.SetTopicAlias, p.TopicAlias, 4)
	eq(t, p.SetMessageExpiryInterval, p.MessageExpiryInterval, 199)
	eq(t, p.SetPayloadFormat, p.PayloadFormat, true)
	eq(t, p.SetResponseTopic, p.ResponseTopic, "a/b/c")
	eq(t, p.SetCorrelationData, p.CorrelationData, []byte("corr"))

	p.AddUserProp("color", "red")
	p.AddSubscriptionID(11)
	if v := p.SubscriptionIDs(); !reflect.DeepEqual(v, []uint32{11}) {
		t.Error("subscriptionIDs", v)
	}

	eq(t, p.SetContentType, p.ContentType, "text/plain")
	eq(t, p.SetPayload, p.Payload, []byte("gopher"))

	var buf bytes.Buffer
	if _, err := p.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}
	t.Logf("\n\n%s\n\n%s\n\n", p, hex.Dump(buf.Bytes()))
}

func Test_QoS(t *testing.T) {
	p := NewPublish()
	eq(t, p.SetQoS, p.QoS, 1)
	eq(t, p.SetQoS, p.QoS, 2)

	if p.SetQoS(3); p.QoS() != 0 {
		t.Error("unexpected qos", p.QoS())
	}
	p.fixed.toggle(QoS1|QoS2, true) // can only happen for incoming packets
	if v := p.QoS(); v != 3 {
		t.Error("unexpected qos", v)
	}
}
