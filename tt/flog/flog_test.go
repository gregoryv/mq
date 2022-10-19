package flog

import (
	"log"
	"os"

	"github.com/gregoryv/mq"
)

func ExampleLogFeature_LogIncoming() {
	log.SetOutput(os.Stdout)
	l := New()
	l.LogLevelSet(LevelInfo)

	p := mq.NewPublish()
	p.SetPayload([]byte("gopher"))
	l.LogIncoming(mq.NoopHandler)(nil, &p)

	// output:
	// in PUBLISH ---- p0 13 bytes
}

func ExampleLogFeature_LogOutgoing() {
	log.SetOutput(os.Stdout)
	l := New()
	l.LogLevelSet(LevelInfo)

	p := mq.NewPublish()
	p.SetPayload([]byte("gopher"))
	l.LogOutgoing(mq.NoopHandler)(nil, &p)

	// output:
	// ut PUBLISH ---- p0 13 bytes
}

func ExampleLogFeature_DumpPacket() {
	log.SetOutput(os.Stdout)
	l := New()
	l.LogLevelSet(LevelDebug)

	p := mq.NewPublish()
	p.SetPayload([]byte("gopher"))
	l.DumpPacket(mq.NoopHandler)(nil, &p)

	// output:
	// 00000000  30 0b 00 00 00 00 06 67  6f 70 68 65 72           |0......gopher|
}

func ExampleLogFeature_PrefixLoggers() {
	log.SetOutput(os.Stdout)
	l := New()
	l.LogLevelSet(1)

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
