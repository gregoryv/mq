package mqtt

import (
	"bytes"
	"testing"

	"github.com/eclipse/paho.golang/packets"
)

func BenchmarkAuth(b *testing.B) {
	b.Run("our", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			p := NewAuth()
			p.AddUserProp("color", "red")
			p.SetReasonCode(ReAuthenticate)
			p.WriteTo(&buf)
			ReadPacket(&buf)
		}
	})
	b.Run("their", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			p := packets.NewControlPacket(packets.AUTH)
			c := p.Content.(*packets.Auth)
			c.ReasonCode = packets.AuthReauthenticate
			c.Properties = &packets.Properties{}
			c.Properties.User = append(
				c.Properties.User, packets.User{"color", "red"},
			)
			p.WriteTo(&buf)
			packets.ReadPacket(&buf)
		}
	})
}

func BenchmarkConnect(b *testing.B) {
	b.Run("our", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			p := NewConnect()
			p.SetKeepAlive(30)
			p.SetClientID("macy")
			p.SetUsername("john.doe")
			p.SetPassword([]byte("secret"))
			p.SetSessionExpiryInterval(30)
			p.WriteTo(&buf)
			ReadPacket(&buf)
		}
	})
	b.Run("their", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
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
			p.WriteTo(&buf)
			packets.ReadPacket(&buf)
		}
	})
}
func BenchmarkConnAck(b *testing.B) {
	b.Run("our", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			p := NewConnAck()
			p.SetAuthMethod("digest")
			p.SetAuthData([]byte("secret"))
			p.WriteTo(&buf)
			ReadPacket(&buf)
		}
	})
	b.Run("their", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			p := packets.NewControlPacket(packets.CONNACK)
			c := p.Content.(*packets.Connack)
			c.Properties = &packets.Properties{}
			c.Properties.AuthMethod = "digest"
			c.Properties.AuthData = []byte("secret")
			p.WriteTo(&buf)
			packets.ReadPacket(&buf)
		}
	})
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
			// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901117
			c.Properties.SubscriptionIdentifier = &subid
			c.Properties.ContentType = "text/plain"
			c.Payload = []byte("gopher")
			p.WriteTo(&buf)
			packets.ReadPacket(&buf)
		}
	})
}
