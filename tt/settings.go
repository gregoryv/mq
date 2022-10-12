package tt

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"github.com/gregoryv/mq"
)

type Settings interface {
	// ReceiverSet configures receiver for any incoming mq.Publish
	// packets. The client handles PacketID reuse.
	ReceiverSet(mq.Handler) error

	LogLevelSet(v LogLevel) error

	// IOSet sets the read writer used for serializing packets from and to.
	// Should be set before calling Run
	IOSet(io.ReadWriter) error
}

type setWrite struct {
	setRead
}

func (s *setWrite) ReceiverSet(v mq.Handler) error {
	s.receiver = v
	return nil
}

func (s *setWrite) LogLevelSet(v LogLevel) error {
	switch v {
	case LogLevelDebug:
		s.info.SetOutput(log.Writer())
		s.debug.SetOutput(log.Writer())

	case LogLevelInfo:
		s.info.SetOutput(log.Writer())
		s.debug.SetOutput(ioutil.Discard)

	case LogLevelNone:
		s.info.SetOutput(ioutil.Discard)
		s.debug.SetOutput(ioutil.Discard)
	}
	return nil
}

func (s *setWrite) IOSet(v io.ReadWriter) error {
	s.wire = v
	return nil
}

type setRead struct {
	*Client
}

func (s *setRead) ReceiverSet(_ mq.Handler) error { return ErrReadOnly }
func (s *setRead) LogLevelSet(_ LogLevel) error   { return ErrReadOnly }
func (s *setRead) IOSet(_ io.ReadWriter) error    { return ErrReadOnly }

var ErrReadOnly = fmt.Errorf("read only")
