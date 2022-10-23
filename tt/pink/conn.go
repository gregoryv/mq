package pink

import "io"

type Conn struct {
	io.Reader // incoming from server
	io.Writer // outgoing to server
}
