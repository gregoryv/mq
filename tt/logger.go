package tt

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"github.com/gregoryv/mq"
)

// NewLogger returns a logger with max id len 11
func NewLogger(v Level) *Logger {
	l := &Logger{
		logLevel: v,
		info:     log.New(ioutil.Discard, "", log.Flags()),
		debug:    log.New(ioutil.Discard, "", log.Flags()),
	}
	switch v {
	case LevelDebug:
		l.info.SetOutput(log.Writer())
		l.debug.SetOutput(log.Writer())

	case LevelInfo:
		l.info.SetOutput(log.Writer())
	}
	l.SetMaxIDLen(11)
	return l
}

type Logger struct {
	logLevel Level
	info     *log.Logger
	debug    *log.Logger

	// client ids
	maxLen uint
}

// SetMaxIDLen configures the logger to trim the client id to number of
// characters. Use 0 to not trim.
func (l *Logger) SetMaxIDLen(max uint) {
	l.maxLen = max
}

func (l *Logger) SetOutput(w io.Writer) {
	l.info.SetOutput(w)
	l.debug.SetOutput(w)
}

// In logs incoming packets and errors from the stack on the
// info level.
func (f *Logger) In(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		f.prefixLoggers(p)
		f.info.Print("in ", p)
		err := next(ctx, p)
		if err != nil {
			f.info.Print(err)
		}
		f.dumpPacket(p)
		return err // return error just incase this middleware is not the first
	}
}

func (f *Logger) Out(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		f.prefixLoggers(p)
		f.info.Print("ut ", p)
		err := next(ctx, p)
		if err != nil {
			f.info.Print(err)
		}
		f.dumpPacket(p)
		return err
	}
}

// prefixLoggers uses the short client id from mq.Connect or
// AssignedClientID from mq.ConnAck as prefix in the loggers.
func (f *Logger) prefixLoggers(p mq.Packet) {
	switch p := p.(type) {
	case *mq.Connect:
		f.setLogPrefix(p.ClientID())

	case *mq.ConnAck:
		if v := p.AssignedClientID(); v != "" {
			f.setLogPrefix(v)
		}
	}
}

func (f *Logger) setLogPrefix(v string) {
	v = tail(v, f.maxLen)
	f.info.SetPrefix(fmt.Sprintf("%s ", v))
	f.debug.SetPrefix(fmt.Sprintf("%s ", v))
}

func (f *Logger) dumpPacket(p mq.Packet) {
	if f.logLevel != LevelDebug {
		return
	}
	var buf bytes.Buffer
	p.WriteTo(&buf)
	f.debug.Print(hex.Dump(buf.Bytes()), "\n")
}

// ----------------------------------------

type Level int

const (
	LevelNone Level = iota
	LevelDebug
	LevelInfo
)

func tail(s string, width uint) string {
	if v := uint(len(s)); v > width {
		return prefix + s[v-width:]
	}
	return s
}

const prefix = "~"
