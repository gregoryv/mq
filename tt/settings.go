package tt

import (
	"io"

	"github.com/gregoryv/mq"
)

type Settings interface {
	InStackSet([]mq.Middleware) error
	OutStackSet([]mq.Middleware) error

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

func (s *writeSettings) InStackSet(v []mq.Middleware) error {
	s.instack = v
	return nil
}

func (s *writeSettings) OutStackSet(v []mq.Middleware) error {
	s.outstack = v
	return nil
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

func (s *readSettings) InStackSet(v []mq.Middleware) error {
	return ErrReadOnly
}
func (s *readSettings) OutStackSet(v []mq.Middleware) error {
	return ErrReadOnly
}

func (s *readSettings) ReceiverSet(_ mq.Handler) error { return ErrReadOnly }
func (s *readSettings) LogLevelSet(_ LogLevel) error   { return ErrReadOnly }
func (s *readSettings) IOSet(_ io.ReadWriter) error    { return ErrReadOnly }
