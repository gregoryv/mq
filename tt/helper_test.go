package tt

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/gregoryv/mq"
)

// Dial returns a test connection to a server used to send responses
// with.
func Dial() (*Conn, *TestServer) {
	fromServer, toClient := io.Pipe()
	toServer := ioutil.Discard
	c := &Conn{
		Reader: fromServer,
		Writer: toServer,
	}
	return c, &TestServer{toClient}
}

type TestServer struct {
	io.Writer
}

func (s *TestServer) Ack(p mq.Packet) {
	switch p := p.(type) {
	case *mq.Subscribe:
		a := mq.NewSubAck()
		a.SetPacketID(p.PacketID())
		a.WriteTo(s)

	case *mq.Connect:
		a := mq.NewConnAck()
		a.WriteTo(s)

	default:
		panic(fmt.Sprintf("TestServer cannot ack %T", p))
	}
}

func (s *TestServer) Pub(qos uint8, topic, payload string) {
	p := mq.NewPublish()
	p.SetQoS(qos)
	p.SetTopicName(topic)
	p.SetPayload([]byte(payload))
	p.WriteTo(s)
}

type Conn struct {
	io.Reader // incoming from server
	io.Writer // outgoing to server
}
