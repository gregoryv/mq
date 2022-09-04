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
	c.setLogPrefix(p.ClientID())
	c.Send(p)
	// check ack
	a, err := mqtt.ReadPacket(c)
	if err != nil {
		return err
	}
	c.Print(a)
	if a, ok := a.(*mqtt.ConnAck); !ok {
		return fmt.Errorf("unexpected ack %T", a)
	} else {
		c.setLogPrefix(a.AssignedClientID())
	}
	return nil
}

func (c *Client) Publish(p *mqtt.Publish) error {
	// todo handle QoS variations
	return c.Send(p)
}

// Send packet to the underlying connection.
func (c *Client) Send(p mqtt.ControlPacket) error {
	// todo handle packet ids I guess

	_, err := p.WriteTo(c)
	if err != nil {
		c.Print(p, err)
		return err
	}
	c.Print(p)
	return nil
}

func (c *Client) setLogPrefix(cid string) {
	switch {
	case cid == "":
		return

	case len(cid) > 16:
		c.Logger.SetPrefix(cid[len(cid)-6:] + " ")

	default:
		c.Logger.SetPrefix(cid + " ")
	}
}
