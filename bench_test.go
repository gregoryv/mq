package mqtt

import (
	"bytes"
	"io"
	"testing"

	"github.com/eclipse/paho.golang/packets"
)

func Benchmark(b *testing.B) {
	cases := []struct {
		name      string
		makeOur   func() io.WriterTo
		makeTheir func() io.WriterTo
	}{
		{"Auth", makeAuth, makeTheirAuth},
		{"Connect", makeConnect, makeTheirConnect},
		{"Publish", makePublish, makeTheirPublish},
	}

	for _, c := range cases {
		b.Run(c.name, func(b *testing.B) {
			b.Run("our", func(b *testing.B) {
				var buf bytes.Buffer
				for i := 0; i < b.N; i++ {
					c.makeOur().WriteTo(&buf)
					ReadPacket(&buf)
				}
			})
			b.Run("their", func(b *testing.B) {
				var buf bytes.Buffer
				for i := 0; i < b.N; i++ {
					c.makeTheir().WriteTo(&buf)
					packets.ReadPacket(&buf)
				}
			})
		})
	}
}

func makeAuth() io.WriterTo {
	p := NewAuth()
	p.AddUserProp("color", "red")
	p.SetReasonCode(ReAuthenticate)
	return &p
}

func makeTheirAuth() io.WriterTo {
	their := packets.NewControlPacket(packets.AUTH)
	c := their.Content.(*packets.Auth)
	c.ReasonCode = packets.AuthReauthenticate
	var (
		p packets.Properties
	)
	c.Properties = &p
	p.User = append(p.User, packets.User{"color", "red"})

	return their
}

// ----------------------------------------

func makeConnect() io.WriterTo {
	p := NewConnect()
	p.SetKeepAlive(30)
	p.SetClientID("macy")
	p.SetUsername("john.doe")
	p.SetPassword([]byte("secret"))
	p.SetSessionExpiryInterval(30)
	return &p
}

func makeTheirConnect() io.WriterTo {
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

func makePublish() io.WriterTo {
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
	return &p
}

func makeTheirPublish() io.WriterTo {
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
