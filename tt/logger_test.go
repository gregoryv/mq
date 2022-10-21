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
	l.In(NoopHandler)(nil, p)

	// output:
	// in PUBLISH ---- p0 a/b 16 bytes
	// 00000000  30 0e 00 03 61 2f 62 00  00 06 67 6f 70 68 65 72  |0...a/b...gopher|
}

func ExampleLogger() {
	log.SetOutput(os.Stdout)
	l := NewLogger(LevelInfo)
	l.SetMaxIDLen(6)
	{
		p := mq.NewConnect()
		p.SetClientID("myclient")
		l.Out(NoopHandler)(nil, &p)
	}
	{
		p := mq.NewConnAck()
		p.SetAssignedClientID("1bbde752-5161-11ed-a94b-675e009b6f46")
		l.In(NoopHandler)(nil, &p)
		l.SetMaxIDLen(0)
		l.In(NoopHandler)(nil, &p)
	}
	// output:
	// ~client ut CONNECT ---- -------- MQTT5 myclient 0s 23 bytes
	// ~9b6f46 in CONNACK ---- -------- 1bbde752-5161-11ed-a94b-675e009b6f46 44 bytes
	// 1bbde752-5161-11ed-a94b-675e009b6f46 in CONNACK ---- -------- 1bbde752-5161-11ed-a94b-675e009b6f46 44 bytes
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
