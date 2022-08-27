package mqtt

import (
	"bytes"
	"testing"

	"github.com/eclipse/paho.golang/packets"
)

func BenchmarkConnect(b *testing.B) {
	b.Run("make", func(b *testing.B) {
		b.Run("our", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_ = makeConnect()
			}
		})
		b.Run("their", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_ = makeTheirConnect()
			}
		})
	})

	// this buf is used in the next Unmarshal, our output is used in
	// both as input

	b.Run("write", func(b *testing.B) {
		b.Run("our", func(b *testing.B) {
			var buf bytes.Buffer
			our := makeConnect()
			for n := 0; n < b.N; n++ {
				buf.Reset()
				our.WriteTo(&buf)
			}
		})

		b.Run("their", func(b *testing.B) {
			var buf bytes.Buffer
			their := makeTheirConnect()
			for n := 0; n < b.N; n++ {
				buf.Reset()
				their.WriteTo(&buf) // to be similar to our
			}
		})
	})

	b.Run("read", func(b *testing.B) {
		var buf bytes.Buffer
		our := makeConnect()
		our.WriteTo(&buf)

		var fh FixedHeader
		fh.ReadFrom(&buf)

		data := make([]byte, buf.Len())
		copy(data, buf.Bytes())

		b.Run("our", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				if _, err := fh.ReadPacket(&buf); err != nil {
					b.Fatal(err)
				}
				buf.Write(data)
			}
		})

		var (
			their = packets.NewControlPacket(packets.CONNECT)
			the   = their.Content.(*packets.Connect)
		)
		b.Run("their", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				if err := the.Unpack(&buf); err != nil {
					b.Fatal(err)
				}
				buf.Write(data)
			}
		})
	})
}

func makeConnect() Connect {
	p := NewConnect()
	p.SetKeepAlive(30)
	p.SetClientID("macy")
	p.SetUsername("john.doe")
	p.SetPassword([]byte("secret"))
	p.SetSessionExpiryInterval(30)
	return p
}

func makeTheirConnect() *packets.ControlPacket {
	p := packets.NewControlPacket(packets.CONNECT)
	c := p.Content.(*packets.Connect)
	c.KeepAlive = 30
	c.ClientID = "macy"
	c.UsernameFlag = true
	c.Username = "john.doe"
	c.PasswordFlag = true
	c.Password = []byte("secret")
	sExpiry := uint32(30)
	c.Properties.SessionExpiryInterval = &sExpiry
	return p
}

// ----------------------------------------

func BenchmarkPublish(b *testing.B) {
	b.Run("make", func(b *testing.B) {
		b.Run("our", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_ = makePublish()
			}
		})

		b.Run("their", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_ = makeTheirPublish()
			}
		})
	})
}

func makePublish() Publish {
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
	return p
}

func makeTheirPublish() *packets.ControlPacket {
	their := packets.NewControlPacket(packets.PUBLISH)
	c := their.Content.(*packets.Publish)
	c.Retain = true
	c.QoS = 1
	c.Duplicate = true
	c.Topic = "topic/name"
	c.PacketID = 1
	var (
		p               packets.Properties
		topicAlias      = uint16(4)
		expInt          = uint32(199)
		pformat         = byte(1)
		correlationData = []byte("corr")
		subid           = 11
	)
	c.Properties = &p
	p.TopicAlias = &topicAlias
	p.MessageExpiry = &expInt
	p.PayloadFormat = &pformat
	p.ResponseTopic = "a/b/c"
	p.CorrelationData = correlationData
	p.User = append(p.User, packets.User{"color", "red"})

	// not fully supported as there can be multiple
	// subscription identifiers
	// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901117
	p.SubscriptionIdentifier = &subid
	p.ContentType = "text/plain"

	c.Payload = []byte("gopher")
	return their
}
