package pakio

import (
	"context"
	"io"
	"io/ioutil"
	"testing"

	"github.com/gregoryv/mq"
)

func TestSender(t *testing.T) {
	s := NewSender(&ClosedConn{})

	ctx := context.Background()
	p := mq.NewConnect()
	if err := s.Send(ctx, &p); err == nil {
		t.Fatal("expect error")
	}
}

// Dial returns a test connection to a server and the server writer
// used to send responses with.
func Dial() (*Conn, io.Writer) {
	r, w := io.Pipe()
	c := &Conn{
		Reader: r,
		Writer: ioutil.Discard, // ignore outgoing packets
	}
	return c, w
}

type Conn struct {
	io.Reader // incoming from server
	io.Writer // outgoing to server
}

// ----------------------------------------

type ClosedConn struct{}

func (c *ClosedConn) Read(_ []byte) (int, error) {
	return 0, io.EOF
}

func (c *ClosedConn) Write(_ []byte) (int, error) {
	return 0, io.EOF
}
