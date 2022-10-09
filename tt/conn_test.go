package tt

import (
	"errors"
	"io"
	"io/ioutil"
)

// Dial returns a test connection to a server and the server writer
// used to send responses with.
func Dial() (*Conn, io.Writer) {
	r, w := io.Pipe()
	c := &Conn{
		Reader: r,
		Writer: ioutil.Discard, // ignore incoming packets
	}
	return c, w
}

type Conn struct {
	io.Reader // from server
	io.Writer // to server
}

func (c *Conn) Read(p []byte) (int, error) {
	n, err := c.Reader.Read(p)
	if errors.Is(io.EOF, err) {
		return n, nil
	}
	return n, err
}
