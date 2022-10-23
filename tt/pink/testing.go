package pink

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/gregoryv/mq"
)

// Dial returns a test connection to a server used to send responses
// with.
func (s *Server) Dial() *Conn {
	fromServer, toClient := io.Pipe()
	toServer := ioutil.Discard
	c := &Conn{
		Reader: fromServer,
		Writer: toServer,
	}
	s.AddConnection(context.Background(), &Conn{
		// Reader: fromClient
		Writer: toClient,
	})
	return c
}

func (s *Server) Ack(p mq.Packet) {
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

func (s *Server) Pub(qos uint8, topic, payload string) {
	p := mq.NewPublish()
	p.SetQoS(qos)
	p.SetTopicName(topic)
	p.SetPayload([]byte(payload))
	p.WriteTo(s)
}
