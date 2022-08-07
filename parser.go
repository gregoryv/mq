package mqtt

import (
	"bytes"
	"fmt"
	"io"
)

func Parse(r io.Reader) (ControlPacket, error) {

	h, err := parseFixedHeader(r)
	if err != nil {
		return nil, fmt.Errorf("ParseControlPacket %w", err)
	}

	var p ControlPacket
	switch {
	case h.Is(CONNECT):
		p = NewConnect()

	default:
		return nil, fmt.Errorf("ParseControlPacket unknown %s", h)
	}
	// read the remaining variable and payload
	l := p.FixedHeader().RemLen()
	rest := make([]byte, l)

	n, _ := r.Read(rest)
	if n != l {
		return p, fmt.Errorf("expected %v bytes read %v, %w", l, n, ErrIncomplete)
	}
	br := bytes.NewReader(rest)
	err = p.Fill(h, br)
	return p, err
}

// parseFixedHeader returns complete or partial header on error
func parseFixedHeader(r io.Reader) (FixedHeader, error) {
	buf := make([]byte, 1)
	header := make(FixedHeader, 0, 5) // max 5

	if _, err := r.Read(buf); err != nil {
		return header, fmt.Errorf("ParseFixedHeader: %w", err)
	}
	header = append(header, buf[0])
	if header.Is(UNDEFINED) {
		return header, ErrTypeUndefined
	}
	v, err := ParseVarInt(r)
	if err != nil {
		return header, err
	}
	header = append(header, NewVarInt(v)...)
	return header, nil
}

var ErrTypeUndefined = fmt.Errorf("type undefined")
