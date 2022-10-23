package tt

import (
	"io"
	"io/ioutil"
)

// Dial returns a test connection to a server used to send responses
// with.
func Dial() (*Conn, io.Writer) {
	fromServer, toClient := io.Pipe()
	toServer := ioutil.Discard
	c := &Conn{
		Reader: fromServer,
		Writer: toServer,
	}
	return c, toClient
}

type Conn struct {
	io.Reader // incoming from server
	io.Writer // outgoing to server
}
