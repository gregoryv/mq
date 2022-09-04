package client

import (
	"fmt"
	"io"
	"log"

	"github.com/gregoryv/mqtt"
)

func NewClient(conn io.ReadWriter) *Client {
	return &Client{
		ReadWriter: conn,
		Logger:     log.New(log.Writer(), "", log.Flags()),
	}
}

type Client struct {
	io.ReadWriter

	*log.Logger
}

// Connect sends the packet and waits for acknoledgement. In the
// future this would be a good place to implement support for
// different auth methods.
func (c *Client) Connect(p *mqtt.Connect) error {
	c.Logger.SetPrefix(p.ClientID() + " ")
	c.Send(p)
	// check that it's acknowledged
	a, err := mqtt.ReadPacket(c)
	if err != nil {
		return err
	}
	c.Print(a)
	if _, ok := a.(*mqtt.ConnAck); !ok {
		return fmt.Errorf("unexpected ack %T", a)
	}
	return nil
}

func (c *Client) Send(p mqtt.ControlPacket) error {
	_, err := p.WriteTo(c)
	if err != nil {
		c.Print(p, err)
		return err
	}
	c.Print(p)
	return nil
}
