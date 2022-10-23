package pink

import "io"

func NewConn(r io.Reader, w io.Writer) *Conn {
	return &Conn{Reader: r, Writer: w}
}

type Conn struct {
	io.Reader // incoming from server
	io.Writer // outgoing to server
}
