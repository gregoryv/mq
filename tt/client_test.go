package tt

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/gregoryv/mq"
	"github.com/gregoryv/mq/tt/flog"
	"github.com/gregoryv/mq/tt/intercept"
)

var _ mq.Client = &Client{}

// thing is anything like an iot device that mostly sends stats to the
// cloud
func TestThingClient(t *testing.T) {
	c := NewBasicClient()
	conn, server := Dial()
	c.IOSet(conn)
	ctx, incoming := runIntercepted(t, c)

	{ // connect mq tt
		p := mq.NewConnect()
		_ = c.Send(ctx, &p)

		ack := mq.NewConnAck()
		ack.WriteTo(server)

		_ = (<-incoming).(*mq.ConnAck)
	}
	{ // publish application message
		p := mq.NewPublish()
		p.SetQoS(1)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		_ = c.Send(ctx, &p)

		ack := mq.NewPubAck()
		ack.SetPacketID(p.PacketID())
		ack.WriteTo(server)
		_ = (<-incoming).(*mq.PubAck)
	}
	{ // disconnect nicely
		p := mq.NewDisconnect()
		if err := c.Send(ctx, &p); err != nil {
			t.Fatal(err)
		}
	}
}

func TestClient_Send(t *testing.T) {
	c := NewBasicClient()
	s := c
	s.IOSet(&ClosedConn{})

	ctx := context.Background()
	p := mq.NewConnect()
	if err := c.Send(ctx, &p); err == nil {
		t.Fatal("expect error")
	}
}

func TestClient_Settings(t *testing.T) {
	c := NewBasicClient()
	s := c
	conn, _ := Dial()

	// before start
	s = c
	if err := s.IOSet(conn); err != nil {
		t.Error(err)
	}
	if err := s.ReceiverSet(nil); err != nil {
		t.Error(err)
	}
	fl := flog.New()
	fl.LogLevelSet(flog.LevelInfo)
	in := []mq.Middleware{fl.LogIncoming}
	if err := s.InStackSet(in); err != nil {
		t.Error(err)
	}

	out := []mq.Middleware{fl.LogOutgoing}
	if err := s.OutStackSet(out); err != nil {
		t.Error(err)
	}

	ctx := context.Background()
	c.Start(ctx)

	// after
	if err := s.IOSet(nil); err == nil {
		t.Error("could set IO after start")
	}
	if err := s.ReceiverSet(nil); err == nil {
		t.Error("could set Receiver after start")
	}
	if err := s.InStackSet(nil); err == nil {
		t.Error("could set InStack after start")
	}
	if err := s.OutStackSet(nil); err == nil {
		t.Error("could set OutStack after start")
	}
}

func TestClient_RunRespectsContextCancel(t *testing.T) {
	c := NewBasicClient()
	s := c
	conn := dialBroker(t)
	s.IOSet(conn)
	var wg sync.WaitGroup
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Millisecond)

	wg.Add(1)
	go func() {
		if err := c.Run(ctx); !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("unexpected error: %T", err)
		}
		wg.Done()
	}()

	wg.Wait()
}

// ----------------------------------------

func runIntercepted(t *testing.T, c *Client) (context.Context, <-chan mq.Packet) {
	r := intercept.New(0)
	c.instack = append([]mq.Middleware{r.Intercept}, c.instack...) // prepend
	ctx, cancel := context.WithCancel(context.Background())
	c.Start(ctx)
	t.Cleanup(cancel)
	return ctx, r.C
}

func newClient(t *testing.T) *Client {
	c := NewBasicClient()
	c.IOSet(dialBroker(t))
	return c
}

func dialBroker(t *testing.T) net.Conn {
	conn, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		t.Skip(err)
		return nil
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}
