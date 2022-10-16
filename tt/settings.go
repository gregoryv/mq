package tt

import (
	"io"

	"github.com/gregoryv/mq"
)

type Settings interface {
	// ReceiverSet configures final handler of the incoming
	// stack. Usually some sort of router to the application logic.
	ReceiverSet(mq.Handler) error

	LogLevelSet(v LogLevel) error

	// IOSet sets the read writer used for serializing packets from and to.
	// Should be set before calling Run
	IOSet(io.ReadWriter) error
}

type writeSettings struct {
	readSettings
}

func (s *writeSettings) ReceiverSet(v mq.Handler) error {
	s.receiver = v
	return nil
}

func (s *writeSettings) LogLevelSet(v LogLevel) error {
	s.flog.LogLevelSet(v)
	return nil

}

func (s *writeSettings) IOSet(v io.ReadWriter) error {
	s.wire = v
	return nil
}

// ----------------------------------------

type readSettings struct {
	*Client
}

func (s *readSettings) ReceiverSet(_ mq.Handler) error { return ErrReadOnly }
func (s *readSettings) LogLevelSet(_ LogLevel) error   { return ErrReadOnly }
func (s *readSettings) IOSet(_ io.ReadWriter) error    { return ErrReadOnly }
