package parser

import (
	"context"
	"fmt"
	"io"

	"github.com/gregoryv/mqtt"
)

func NewParser(r io.Reader) *Parser {
	return &Parser{
		r: r,
	}
}

type Parser struct {
	r io.Reader
}

func (p *Parser) Parse(ctx Context, c chan<- *mqtt.ControlPacket) error {
	for {
		next, err := ParseControlPacket(ctx, p.r)
		if err != nil {
			return err
		}
		c <- next
	}
	return fmt.Errorf(": todo")
}

func ParseControlPacket(_ Context, r io.Reader) (*mqtt.ControlPacket, error) {
	h, err := ParseFixedHeader(r)
	if err != nil {
		return nil, err
	}

	cp := &mqtt.ControlPacket{
		FixedHeader: h,
	}
	return cp, nil
}

type Context = context.Context

func ParseFixedHeader(r io.Reader) (mqtt.FixedHeader, error) {
	buf := make([]byte, 1)
	header := make([]byte, 0, 5) // max 5

	if _, err := r.Read(buf); err != nil {
		return header, err
	}
	header = append(header, buf[0])
	v, err := mqtt.ParseVarInt(r)
	if err != nil {
		return header, err
	}
	header = append(header, mqtt.NewVarInt(v)...)
	return header, nil
}
