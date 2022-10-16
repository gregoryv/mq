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

func NewLogFeature() *LogFeature {
	return &LogFeature{
		logLevel: LogLevelNone,
		info:     log.New(log.Writer(), "", log.Flags()),
		debug:    log.New(log.Writer(), "", log.Flags()),
	}
}

type LogFeature struct {
	logLevel LogLevel
	info     *log.Logger
	debug    *log.Logger
}

func (f *LogFeature) LogLevelSet(v LogLevel) {
	switch v {
	case LogLevelDebug:
		f.info.SetOutput(log.Writer())
		f.debug.SetOutput(log.Writer())

	case LogLevelInfo:
		f.info.SetOutput(log.Writer())
		f.debug.SetOutput(ioutil.Discard)

	case LogLevelNone:
		f.info.SetOutput(ioutil.Discard)
		f.debug.SetOutput(ioutil.Discard)
	}
	f.logLevel = v
}

func (f *LogFeature) prefixLoggers(next mq.Handler) mq.Handler {
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

func (f *LogFeature) LogIncoming(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		f.info.Print("in ", p)
		return next(ctx, p)
	}
}

func (f *LogFeature) DumpPacket(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		if f.logLevel == LogLevelDebug {
			var buf bytes.Buffer
			p.WriteTo(&buf)
			f.debug.Print(hex.Dump(buf.Bytes()), "\n")
		}
		return next(ctx, p)
	}
}

func (f *LogFeature) logOutgoing(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		f.info.Print("ut ", p)
		return next(ctx, p)
	}
}

func (f *LogFeature) setLogPrefix(cid string) {
	f.info.SetPrefix(fmt.Sprintf("%s ", cid))
	f.debug.SetPrefix(fmt.Sprintf("%s ", cid))
}