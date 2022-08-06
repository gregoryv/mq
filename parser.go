package mqtt

import (
	"errors"
	"fmt"
	"io"
)

func Parse(r io.Reader) (p ControlPacket, err error) {
	h, err := parseFixedHeader(r)
	if err != nil {
		return nil, fmt.Errorf("ParseControlPacket %w", err)
	}

	switch {
	case h.Is(CONNECT):
		p = NewConnect()

	default:
		err = fmt.Errorf("ParseControlPacket unknown %s", h)
		return
	}
	// read the remaining variable and payload
	rest := make([]byte, p.FixedHeader().RemLen())
	_, err = r.Read(rest)
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	err = p.Fill(h, rest)
	return
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
