package tt

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/gregoryv/mq"
)

func NewLogger(v Level) *Logger {
	f := &Logger{
		logLevel: v,
		info:     log.New(ioutil.Discard, "", log.Flags()),
		debug:    log.New(ioutil.Discard, "", log.Flags()),
	}
	switch v {
	case LevelDebug:
		f.info.SetOutput(log.Writer())
		f.debug.SetOutput(log.Writer())

	case LevelInfo:
		f.info.SetOutput(log.Writer())
	}
	return f
}

type Logger struct {
	logLevel Level
	info     *log.Logger
	debug    *log.Logger
}

// PrefixLoggers uses the short client id from mq.Connect or
// AssignedClientID from mq.ConnAck as prefix in the loggers.
func (f *Logger) PrefixLoggers(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		switch p := p.(type) {
		case *mq.Connect:
			f.setLogPrefix(p.ClientIDShort())

		case *mq.ConnAck:
			if p.AssignedClientID() != "" {
				f.setLogPrefix(p.AssignedClientID())
			}
		}
		return next(ctx, p)
	}
}

// LogIncoming logs incoming packets and errors from the stack on the
// info level.
func (f *Logger) LogIncoming(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		f.info.Print("in ", p)
		err := next(ctx, p)
		if err != nil {
			f.info.Print(err)
		}
		return err // return error just incase this middleware is not the first
	}
}

func (f *Logger) DumpPacket(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		if f.logLevel == LevelDebug {
			var buf bytes.Buffer
			p.WriteTo(&buf)
			f.debug.Print(hex.Dump(buf.Bytes()), "\n")
		}
		return next(ctx, p)
	}
}

func (f *Logger) LogOutgoing(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		f.info.Print("ut ", p)
		err := next(ctx, p)
		if err != nil {
			f.info.Print(err)
		}
		return err
	}
}

func (f *Logger) setLogPrefix(cid string) {
	f.info.SetPrefix(fmt.Sprintf("%s ", cid))
	f.debug.SetPrefix(fmt.Sprintf("%s ", cid))
}

// ----------------------------------------

type Level int

const (
	LevelNone Level = iota
	LevelDebug
	LevelInfo
)
