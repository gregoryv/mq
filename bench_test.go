package mqtt

import (
	"bytes"
	"io"
	"testing"

	"github.com/eclipse/paho.golang/packets"
)

func BenchmarkAuth(b *testing.B) {
	b.Run("make", func(b *testing.B) {
		b.Run("our", benchMake(b, func() { _ = makeAuth() }))
		b.Run("their", benchMake(b, func() { _ = makeTheirAuth() }))
	})
	b.Run("write", func(b *testing.B) {
		our := makeAuth()
		b.Run("our", benchWriteTo(b, &our))
		b.Run("their", benchWriteTo(b, makeTheirAuth()))
	})
	b.Run("read", func(b *testing.B) {
		// prepare data whith everything after the fixed header
		p := makeAuth()
		fh, data := prepareRead(&p)
		b.Run("our", benchReadRemaining(data, fh))
		their := packets.NewControlPacket(packets.AUTH)
		b.Run("their", benchUnpack(data, their.Content.(*packets.Auth)))
	})
}

func makeAuth() Auth {
	p := NewAuth()
	p.AddUserProp("color", "red")
	p.SetReasonCode(ReAuthenticate)
	return p
}

func makeTheirAuth() *packets.ControlPacket {
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

func BenchmarkConnect(b *testing.B) {
	b.Run("make", func(b *testing.B) {
		b.Run("our", benchMake(b, func() { _ = makeConnect() }))
		b.Run("their", benchMake(b, func() { _ = makeTheirConnect() }))
	})

	b.Run("write", func(b *testing.B) {
		our := makeConnect()
		b.Run("our", benchWriteTo(b, &our))
		b.Run("their", benchWriteTo(b, makeTheirConnect()))
	})

	b.Run("read", func(b *testing.B) {
		// prepare data whith everything after the fixed header
		p := makeConnect()
		fh, data := prepareRead(&p)
		b.Run("our", benchReadRemaining(data, fh))
		their := packets.NewControlPacket(packets.CONNECT)
		b.Run("their", benchUnpack(data, their.Content.(*packets.Connect)))
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
		b.Run("our", benchMake(b, func() { _ = makePublish() }))
		b.Run("their", benchMake(b, func() { _ = makeTheirPublish() }))
	})

	b.Run("write", func(b *testing.B) {
		our := makePublish()
		b.Run("our", benchWriteTo(b, &our))
		b.Run("their", benchWriteTo(b, makeTheirPublish()))
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

// ----------------------------------------

func benchWriteTo(b *testing.B, p io.WriterTo) func(b *testing.B) {
	return func(b *testing.B) {
		var buf bytes.Buffer
		for n := 0; n < b.N; n++ {
			buf.Reset()
			p.WriteTo(&buf)
		}
	}
}

func benchMake(b *testing.B, make func()) func(b *testing.B) {
	return func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			make()
		}
	}
}

func prepareRead(p ControlPacket) (*FixedHeader, []byte) {
	var buf bytes.Buffer
	p.WriteTo(&buf)

	var fh FixedHeader
	fh.ReadFrom(&buf)

	data := make([]byte, buf.Len())
	copy(data, buf.Bytes())
	return &fh, data
}

func benchReadRemaining(data []byte, fh *FixedHeader) func(b *testing.B) {
	var buf bytes.Buffer
	buf.Write(data)
	return func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			if _, err := fh.ReadRemaining(&buf); err != nil {
				b.Fatal(err)
			}
			buf.Write(data)
		}
	}
}

func benchUnpack(data []byte, p interface{ Unpack(*bytes.Buffer) error }) func(b *testing.B) {
	var buf bytes.Buffer
	buf.Write(data)
	return func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			if err := p.Unpack(&buf); err != nil {
				b.Fatal(err)
			}
			buf.Write(data)
		}
	}
}
