package tt

import (
	"fmt"

	"github.com/gregoryv/mq"
)

type Packet struct {
	client *Client
	ack    interface{}
	*mq.Publish
}

func (p *Packet) Client() mq.Client { return p.client }
func (p *Packet) IsAck() bool       { return p.ack != nil }

func (p *Packet) String() string {
	if p.IsAck() {
		return p.ack.(fmt.Stringer).String()
	}
	return p.Publish.String()
}
