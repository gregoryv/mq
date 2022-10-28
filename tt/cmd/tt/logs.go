package main

import (
	"io"

	"github.com/gregoryv/mq/tt"
)

// NextLogWriter if set is use in next call to NewLogger.
var NextLogWriter io.Writer

// NewLogger returns a logger with max id len 11
func NewLogger(v tt.Level) *tt.Logger {
	l := tt.NewLogger(v)
	if v := NextLogWriter; v != nil {
		l.SetOutput(v)
		NextLogWriter = nil
	}
	return l
}
