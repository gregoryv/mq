package tt

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/gregoryv/mq"
)

var _ mq.Client = &Client{}

// thing is anything like an iot device that mostly sends stats to the
// cloud
func TestThingClient(t *testing.T) {
	c := NewBasicClient()
	conn, server := Dial()
	c.Settings().IOSet(conn)
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

func TestAppClient(t *testing.T) {
	c := newClient(t)
	//c.LogLevelSet(LogLevelDebug)
	ctx, incoming := runIntercepted(t, c)

	{ // connect mq tt
		p := mq.NewConnect()
		_ = c.Send(ctx, &p)
		_ = (<-incoming).(*mq.ConnAck)
	}
	{ // subscribe
		p := mq.NewSubscribe()
		p.AddFilter("a/b", mq.OptQoS1)
		_ = c.Send(ctx, &p)
		_ = (<-incoming).(*mq.SubAck)
	}
	{ // publish application message
		p := mq.NewPublish()
		p.SetQoS(1)
		p.SetTopicName("a/b")
		p.SetPayload([]byte("gopher"))
		_ = c.Send(ctx, &p)
		_ = (<-incoming).(*mq.PubAck)
		_ = (<-incoming).(*mq.Publish)
	}
	{ // disconnect nicely
		p := mq.NewDisconnect()
		_ = c.Send(ctx, &p)
	}
}

func TestClient_Send(t *testing.T) {
	c := NewBasicClient()
	s := c.Settings()
	s.IOSet(&ClosedConn{})

	ctx := context.Background()
	p := mq.NewConnect()
	if err := c.Send(ctx, &p); err == nil {
		t.Fatal("expect error")
	}
}

func TestClient_Settings(t *testing.T) {
	c := NewBasicClient()
	s := c.Settings()
	conn, _ := Dial()
	s.IOSet(conn)
	ctx := context.Background()
	c.Start(ctx)

	s = c.Settings()
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
	s := c.Settings()
	conn := dialBroker(t)
	s.IOSet(conn)
	var wg sync.WaitGroup
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Millisecond)

	wg.Add(1)
	go func() {
		_ = c.Run(ctx)
		wg.Done()
	}()

	wg.Wait()
}

// ----------------------------------------

func runIntercepted(t *testing.T, c *Client) (context.Context, <-chan mq.Packet) {
	r := NewInterceptor(0)
	c.instack = append([]mq.Middleware{r.intercept}, c.instack...) // prepend
	ctx, cancel := context.WithCancel(context.Background())
	c.Start(ctx)
	t.Cleanup(cancel)
	return ctx, r.C
}

// todo move as own feature
func NewInterceptor(max int) *Interceptor {
	return &Interceptor{
		C: make(chan mq.Packet, max),
	}
}

type Interceptor struct {
	C chan mq.Packet
}

func (r *Interceptor) intercept(next mq.Handler) mq.Handler {
	return func(ctx context.Context, p mq.Packet) error {
		select {
		case r.C <- p: // if anyone is interested
		case <-time.After(10 * time.Millisecond):
		}
		return next(ctx, p)
	}
}

func newClient(t *testing.T) *Client {
	c := NewBasicClient()
	s := c.Settings()
	s.IOSet(dialBroker(t))
	return c
}

func dialBroker(t *testing.T) net.Conn {
	conn, err := net.Dial("tcp", "127.0.0.1:1883")
	if err != nil {
		t.Log("no broker, did you run docker-compose up?")
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}
