package tt

import (
	"bytes"
	"context"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"

	"github.com/gregoryv/mq"
)

func init() {
	log.SetFlags(0) // quiet by default
}

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

// In logs incoming packets. Log prefix is based on
// mq.ConnAck.AssignedClientID.
func (f *Logger) In(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		if p, ok := p.(*mq.ConnAck); ok {
			if v := p.AssignedClientID(); v != "" {
				f.SetLogPrefix(v)
			}
		}
		// double spaces to align in/out. Usually this is not advised
		// but in here it really does aid when scanning for patterns
		// of packets.
		f.info.Print("in  ", p)
		err := next(ctx, p)
		if err != nil {
			f.info.Print(err)
		}
		if f.logLevel == LevelDebug {
			f.dumpPacket(p)
		}
		// return error just incase this middleware is not the first
		return err
	}
}

// Out logs outgoing packets. Log prefix is based on
// mq.Connect.ClientID.
func (f *Logger) Out(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		if p, ok := p.(*mq.Connect); ok {
			f.SetLogPrefix(p.ClientID())
		}
		f.info.Print("out ", p)
		err := next(ctx, p)
		if err != nil {
			f.info.Print(err)
		}
		if f.logLevel == LevelDebug {
			f.dumpPacket(p)
		}
		return err
	}
}

func (f *Logger) SetLogPrefix(v string) {
	v = newPrefix(v, f.maxLen)
	f.info.SetPrefix(v + " ")
	f.debug.SetPrefix(v + " ")
}

func (f *Logger) dumpPacket(p mq.Packet) {
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

func newPrefix(s string, width uint) string {
	if v := uint(len(s)); v > width {
		return prefixStr + s[v-width:]
	}
	return s
}

const prefixStr = "~"
