package mq

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/eclipse/paho.golang/packets"
)

func ExamplePublish_stringMalformed() {
	fmt.Println(Pub(0, "", "gopher"))
	fmt.Println(Pub(3, "a/b", "gopher"))
	fmt.Println(Pub(1, "a/b", "gopher"))
	// output:
	// PUBLISH ---- p0  13 bytes, malformed! empty topic name
	// PUBLISH -!!- p0 a/b 16 bytes, malformed! invalid QoS
	// PUBLISH --1- p0 a/b 18 bytes, malformed! empty packet ID
}

func ExamplePublish_String() {
	p := Pub(2, "a/b/1", "gopher")
	p.SetPacketID(3)
	p.SetRetain(true)
	p.SetCorrelationData([]byte("1111-222222-3333333"))
	fmt.Println(p)
	fmt.Print(DocumentFlags(p))
	// output:
	// PUBLISH -2-r p3 a/b/1 1111-222222-3333333 42 bytes
	//         3210 PacketID Topic [CorrelationData] Size
	//
	// 3 d   Duplicate
	// 2 2|! QoS
	// 1 1|! QoS
	// 0 r   Retain
}

func ExamplePublish_StringWithoutCorrelation() {
	p := Pub(0, "a/b/1", "gopher")
	fmt.Println(p)
	// output:
	// PUBLISH ---- p0 a/b/1 18 bytes
}

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
	eq(t, p.SetContentType, p.ContentType, "text/plain")
	eq(t, p.SetPayload, p.Payload, []byte("gopher"))

	p.AddUserProp("color", "red")
	p.AddSubscriptionID(11)
	if v := p.SubscriptionIDs(); !reflect.DeepEqual(v, []uint32{11}) {
		t.Error("subscriptionIDs", v)
	}

	testControlPacket(t, p)
}

func Test_QoS(t *testing.T) {
	p := NewPublish()
	eq(t, p.SetQoS, p.QoS, 1)
	eq(t, p.SetQoS, p.QoS, 2)
	eq(t, p.SetQoS, p.QoS, 3)

	if p.SetQoS(9); p.QoS() != 0 {
		t.Error("unexpected qos", p.QoS())
	}
	p.fixed.toggle(QoS3, true) // can only happen for incoming packets
	if v := p.QoS(); v != 3 {
		t.Error("unexpected qos", v)
	}
}

func TestComparePublish(t *testing.T) {
	our := NewPublish()
	// theirs is divided into a wrapping ControlPacket and content
	their := packets.NewControlPacket(packets.PUBLISH)
	the := their.Content.(*packets.Publish)

	our.SetTopicName("topic/")
	the.Topic = "topic/"

	//our.SetRetain(true)
	// bug in pahos, Publish.WriteTo sets the flags, though it's never
	// used if new control packet is created with packets.NewControlPacket
	//the.Retain = true
	//the.Duplicate = true

	// no reason to continue the comparison until the above bug is fixed
	compare(t, our, their)
}

func BenchmarkPublish(b *testing.B) {
	b.Run("our", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			p := NewPublish()
			p.SetRetain(true)
			p.SetQoS(1)
			p.SetDuplicate(true)
			p.SetTopicName("topic/name")
			p.SetPacketID(1)
			p.SetTopicAlias(4)
			p.SetMessageExpiryInterval(199)
			p.SetPayloadFormat(true)
			p.SetResponseTopic("a/b/c")
			p.SetCorrelationData([]byte("corr"))
			p.AddUserProp("color", "red")
			p.AddSubscriptionID(11)
			p.SetContentType("text/plain")
			p.SetPayload([]byte("gopher"))
			p.WriteTo(&buf)
			ReadPacket(&buf)
		}
	})
	b.Run("their", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			p := packets.NewControlPacket(packets.PUBLISH)
			c := p.Content.(*packets.Publish)
			c.Retain = true
			c.QoS = 1
			c.Duplicate = true
			c.Topic = "topic/name"
			c.PacketID = 1
			var (
				topicAlias      = uint16(4)
				expInt          = uint32(199)
				pformat         = byte(1)
				correlationData = []byte("corr")
				subid           = 11
			)
			c.Properties = &packets.Properties{}
			c.Properties.TopicAlias = &topicAlias
			c.Properties.MessageExpiry = &expInt
			c.Properties.PayloadFormat = &pformat
			c.Properties.ResponseTopic = "a/b/c"
			c.Properties.CorrelationData = correlationData
			c.Properties.User = append(
				c.Properties.User, packets.User{"color", "red"},
			)

			// not fully supported as there can be multiple
			// subscription identifiers
			// https://docs.oasis-open.org/mq/mq/v5.0/os/mq-v5.0-os.html#_Toc3901117
			c.Properties.SubscriptionIdentifier = &subid
			c.Properties.ContentType = "text/plain"
			c.Payload = []byte("gopher")
			p.WriteTo(&buf)
			packets.ReadPacket(&buf)
		}
	})
}
