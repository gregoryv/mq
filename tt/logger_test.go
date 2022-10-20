package tt

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gregoryv/mq"
)

func ExampleLogger_In() {
	log.SetOutput(os.Stdout)
	l := NewLogger(LevelInfo)

	p := mq.Pub(0, "a/b", "gopher")
	l.In(NoopHandler)(nil, p)

	// output:
	// in PUBLISH ---- p0 a/b 16 bytes
}

func ExampleLogger_Out() {
	log.SetOutput(os.Stdout)
	l := NewLogger(LevelInfo)

	p := mq.Pub(0, "a/b", "gopher")
	l.Out(NoopHandler)(nil, p)

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
		l.Out(NoopHandler)(nil, &p)
	}
	{
		p := mq.NewConnAck()
		p.SetAssignedClientID("1-12-123")
		l.In(NoopHandler)(nil, &p)
	}

	// output:
	// myclient ut CONNECT ---- -------- MQTT5 myclient 0s 23 bytes
	// 1-12-123 in CONNACK ---- -------- 1-12-123 16 bytes
}

func ExampleLogger_errors() {
	log.SetOutput(os.Stdout)
	l := NewLogger(LevelInfo)

	p := mq.Pub(0, "a/b", "gopher")
	broken := func(context.Context, mq.Packet) error {
		return fmt.Errorf("broken")
	}
	l.In(broken)(nil, p)
	l.Out(broken)(nil, p)
	// output:
	// in PUBLISH ---- p0 a/b 16 bytes
	// broken
	// ut PUBLISH ---- p0 a/b 16 bytes
	// broken
}
