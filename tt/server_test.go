package tt

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/gregoryv/mq"
)

// Dial returns a test connection to a server and the server writer
// used to send responses with.
func Dial() (*Conn, *TestServer) {
	r, w := io.Pipe()
	c := &Conn{
		Reader: r,
		Writer: ioutil.Discard, // ignore outgoing packets
	}
	return c, &TestServer{w}
}

type Conn struct {
	io.Reader // incoming from server
	io.Writer // outgoing to server
}

type TestServer struct {
	io.Writer
}

func (t *TestServer) Ack(p mq.Packet) {
	switch p := p.(type) {
	case *mq.Subscribe:
		a := mq.NewSubAck()
		a.SetPacketID(p.PacketID())
		a.WriteTo(t)
	case *mq.Connect:
		a := mq.NewConnAck()
		a.WriteTo(t)
	default:
		panic(fmt.Sprintf("TestServer cannot ack %T", p))
	}
}

func (t *TestServer) Pub(qos uint8, topic, payload string) {
	p := mq.NewPublish()
	p.SetQoS(qos)
	p.SetTopicName(topic)
	p.SetPayload([]byte(payload))
	p.WriteTo(t)
}
