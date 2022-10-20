package tt

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gregoryv/mq"
)

func ExampleLogger_LogIncoming() {
	log.SetOutput(os.Stdout)
	l := NewLogger(LevelInfo)

	p := mq.NewPublish()
	p.SetPayload([]byte("gopher"))
	l.LogIncoming(mq.NoopHandler)(nil, &p)

	// output:
	// in PUBLISH ---- p0 13 bytes
}

func ExampleLogger_LogOutgoing() {
	log.SetOutput(os.Stdout)
	l := NewLogger(LevelInfo)

	p := mq.NewPublish()
	p.SetPayload([]byte("gopher"))
	l.LogOutgoing(mq.NoopHandler)(nil, &p)

	// output:
	// ut PUBLISH ---- p0 13 bytes
}

func ExampleLogger_DumpPacket() {
	log.SetOutput(os.Stdout)
	l := NewLogger(LevelDebug)

	p := mq.NewPublish()
	p.SetPayload([]byte("gopher"))
	l.DumpPacket(mq.NoopHandler)(nil, &p)

	// output:
	// 00000000  30 0b 00 00 00 00 06 67  6f 70 68 65 72           |0......gopher|
}

func ExampleLogger_PrefixLoggers() {
	log.SetOutput(os.Stdout)
	l := NewLogger(LevelDebug)

	{
		p := mq.NewConnect()
		p.SetClientID("myclient")
		l.PrefixLoggers(mq.NoopHandler)(nil, &p)
	}
	{
		p := mq.NewPublish()
		p.SetPayload([]byte("gopher"))

		l.LogOutgoing(mq.NoopHandler)(nil, &p)
	}
	{
		p := mq.NewConnAck()
		p.SetAssignedClientID("123456789-123456789-123456789")
		l.PrefixLoggers(mq.NoopHandler)(nil, &p)
	}
	{
		p := mq.NewPublish()
		p.SetPayload([]byte("gopher"))
		l.LogIncoming(mq.NoopHandler)(nil, &p)
	}
	// output:
	// myclient ut PUBLISH ---- p0 13 bytes
	// 123456789-123456789-123456789 in PUBLISH ---- p0 13 bytes
}

func ExampleLogger_errors() {
	log.SetOutput(os.Stdout)
	l := NewLogger(LevelInfo)

	p := mq.NewPublish()
	p.SetPayload([]byte("gopher"))
	broken := func(context.Context, mq.Packet) error {
		return fmt.Errorf("broken")
	}
	l.LogIncoming(broken)(nil, &p)
	l.LogOutgoing(broken)(nil, &p)
	// output:
	// in PUBLISH ---- p0 13 bytes
	// broken
	// ut PUBLISH ---- p0 13 bytes
	// broken
}
