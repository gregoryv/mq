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

	p := mq.Pub(0, "a/b", "gopher")
	l.LogIncoming(NoopHandler)(nil, p)

	// output:
	// in PUBLISH ---- p0 a/b 16 bytes
}

func ExampleLogger_LogOutgoing() {
	log.SetOutput(os.Stdout)
	l := NewLogger(LevelInfo)

	p := mq.Pub(0, "a/b", "gopher")
	l.LogOutgoing(NoopHandler)(nil, p)

	// output:
	// ut PUBLISH ---- p0 a/b 16 bytes
}

func ExampleLogger_DumpPacket() {
	log.SetOutput(os.Stdout)
	l := NewLogger(LevelDebug)

	p := mq.Pub(0, "a/b", "gopher")
	l.DumpPacket(NoopHandler)(nil, p)

	// output:
	// 00000000  30 0e 00 03 61 2f 62 00  00 06 67 6f 70 68 65 72  |0...a/b...gopher|
}

func ExampleLogger_PrefixLoggers() {
	log.SetOutput(os.Stdout)
	l := NewLogger(LevelDebug)

	{
		p := mq.NewConnect()
		p.SetClientID("myclient")
		l.PrefixLoggers(NoopHandler)(nil, &p)
	}
	{
		p := mq.Pub(0, "a/b", "gopher")
		l.LogOutgoing(NoopHandler)(nil, p)
	}
	{
		p := mq.NewConnAck()
		p.SetAssignedClientID("123456789-123456789-123456789")
		l.PrefixLoggers(NoopHandler)(nil, &p)
	}
	{
		p := mq.Pub(0, "a/c", "gopher")
		l.LogIncoming(NoopHandler)(nil, p)
	}
	// output:
	// myclient ut PUBLISH ---- p0 a/b 16 bytes
	// 123456789-123456789-123456789 in PUBLISH ---- p0 a/c 16 bytes
}

func ExampleLogger_errors() {
	log.SetOutput(os.Stdout)
	l := NewLogger(LevelInfo)

	p := mq.Pub(0, "a/b", "gopher")
	broken := func(context.Context, mq.Packet) error {
		return fmt.Errorf("broken")
	}
	l.LogIncoming(broken)(nil, p)
	l.LogOutgoing(broken)(nil, p)
	// output:
	// in PUBLISH ---- p0 a/b 16 bytes
	// broken
	// ut PUBLISH ---- p0 a/b 16 bytes
	// broken
}
