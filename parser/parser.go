package parser

import (
	"context"
	"fmt"
	"io"
	"log"

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
loop:
	for {
		next, err := ParseControlPacket(ctx, p.r)
		if err != nil {
			return err
		}
		log.Print("DEBUG ", next)
		// The parsing can only be interrupted between two packet
		// reads or if the reader is closed.
		select {
		case c <- next:
			// coverage thing
			_ = 1

		case <-ctx.Done():
			break loop
		}
	}
	return ctx.Err()
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
	header := make(mqtt.FixedHeader, 0, 5) // max 5

	if _, err := r.Read(buf); err != nil {
		return header, err
	}
	header = append(header, buf[0])
	if header.Is(mqtt.UNDEFINED) {
		return nil, MalformedPacket(
			fmt.Sprintf("undefined %v control packet type", mqtt.UNDEFINED),
		)
	}
	v, err := mqtt.ParseVarInt(r)
	if err != nil {
		return header, err
	}
	header = append(header, mqtt.NewVarInt(v)...)
	return header, nil
}

type MalformedPacket string

func (e MalformedPacket) Error() string {
	return string(e)
}
