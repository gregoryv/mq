package mqtt

import (
	"context"
	"fmt"
	"io"
)

func NewParser(r io.Reader) *Parser {
	return &Parser{
		r: r,
	}
}

type Parser struct {
	r io.Reader
}

func (p *Parser) Parse(ctx Context, c chan<- ControlPacket) error {
loop:
	for {
		next, err := ParseControlPacket(ctx, p.r)
		if err != nil {
			debug.Println(err)
			return err
		}
		// The parsing can only be interrupted between two packet
		// reads or if the reader is closed.
		select {
		case c <- next:
			_ = 1 // coverage thing

		case <-ctx.Done():
			break loop
		}
	}
	return ctx.Err()
}

func ParseControlPacket(_ Context, r io.Reader) (ControlPacket, error) {
	h, err := ParseFixedHeader(r)
	if err != nil {
		return nil, fmt.Errorf("ParseControlPacket %w", err)
	}

	var cp ControlPacket
	switch {
	case h.Is(CONNECT):
		cp = &Connect{fixed: h}
	default:
		err = fmt.Errorf("ParseControlPacket unknown %s", h)
	}
	return cp, err
}

type Context = context.Context

func ParseFixedHeader(r io.Reader) (FixedHeader, error) {
	buf := make([]byte, 1)
	header := make(FixedHeader, 0, 5) // max 5

	if _, err := r.Read(buf); err != nil {
		return header, fmt.Errorf("ParseFixedHeader: %w", err)
	}
	header = append(header, buf[0])
	if header.Is(UNDEFINED) {
		return nil, TypeError(
			"ParseFixedHeader: type " + typeNames[UNDEFINED],
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
