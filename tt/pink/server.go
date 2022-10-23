package pink

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/gregoryv/mq"
)

func NewServer() *Server {
	return &Server{
		bind:          ":", // random
		acceptTimeout: time.Millisecond,
		clients:       make(map[string]io.ReadWriter),
	}
}

type Server struct {
	bind          string
	acceptTimeout time.Duration

	io.Writer

	clients map[string]io.ReadWriter
}

// Dial returns a test connection to a server used to send responses
// with.
func (s *Server) Dial() *Conn {
	fromServer, toClient := io.Pipe()
	toServer := ioutil.Discard
	c := &Conn{
		Reader: fromServer,
		Writer: toServer,
	}
	s.AddConnection(&Conn{
		// Reader: fromClient
		Writer: toClient,
	})
	return c
}

func (s *Server) AddConnection(v io.ReadWriter) {
	// todo create in/out queues for each connection
	s.Writer = v
}

func (s *Server) Run(ctx context.Context) error {
	l, err := net.Listen("tcp", s.bind)
	if err != nil {
		return err
	}
	defer l.Close()

	c := make(chan error, 0)

	go func() {
	loop:
		for {
			// Wait for a connection.
			if l, ok := l.(interface{ SetDeadline(time.Time) error }); ok {
				l.SetDeadline(time.Now().Add(s.acceptTimeout))
			}
			conn, err := l.Accept()
			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					continue loop
				}
				c <- err
			}
			// todo go handle connection
			_ = conn
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-c:
		return err
	}
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
