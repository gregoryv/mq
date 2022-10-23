package tt

import (
	"io"

	"github.com/gregoryv/mq/tt/pink"
)

// Dial returns a test connection to a non running server.
func Dial() (io.ReadWriter, *pink.Server) {
	s := pink.NewServer()
	c := s.Dial()
	return c, s
}
