package mqtt

import (
	"context"
	"fmt"
	"io"
	"log"
)

func NewParser(r io.Reader) *Parser {
	return &Parser{
		r: r,
	}
}

type Parser struct {
	r io.Reader
}

func (p *Parser) Parse(ctx Context, c chan<- *ControlPacket) error {
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

func ParseControlPacket(_ Context, r io.Reader) (*ControlPacket, error) {
	h, err := ParseFixedHeader(r)
	if err != nil {
		return nil, err
	}

	cp := &ControlPacket{
		FixedHeader: h,
	}
	return cp, nil
}

type Context = context.Context

func ParseFixedHeader(r io.Reader) (FixedHeader, error) {
	buf := make([]byte, 1)
	header := make(FixedHeader, 0, 5) // max 5

	if _, err := r.Read(buf); err != nil {
		return header, err
	}
	header = append(header, buf[0])
	if header.Is(UNDEFINED) {
		return nil, TypeError(
			fmt.Sprintf("undefined %v control packet type", UNDEFINED),
		)
	}
	v, err := ParseVarInt(r)
	if err != nil {
		return header, err
	}
	header = append(header, NewVarInt(v)...)
	return header, nil
}

type TypeError string

func (e TypeError) Error() string {
	return string(e)
}
