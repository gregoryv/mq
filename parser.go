package mqtt

import (
	"fmt"
	"io"
)

func NewParser() *Parser {
	return &Parser{}
}

type Parser struct{}

func (p *Parser) Parse(r io.Reader) (ControlPacket, error) {
	h, err := parseFixedHeader(r)
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

func parseFixedHeader(r io.Reader) (FixedHeader, error) {
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
