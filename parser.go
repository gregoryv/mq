package mqtt

import (
	"fmt"
	"io"
)

// Parse parses a control packet from the given reader. It assumes the
// first byte is a fixed header.
func Parse(r io.Reader) (interface{}, error) {
	// read fixed header
	h, err := parseFixedHeader(r)
	if err != nil {
		return nil, fmt.Errorf("ParseControlPacket %w", err)
	}
	// read the remaining data for this control packet
	exp := h.RemLen()
	remaining := make([]byte, exp)
	n, _ := r.Read(remaining)
	// fail if we didn't get all the remaining bytes
	if n != exp {
		return nil, fmt.Errorf(
			"expected %v bytes read %v, %w", exp, n, ErrIncomplete,
		)
	}

	switch {
	case h.Is(CONNECT):
		p := NewConnect()
		return p, p.UnmarshalBinary(remaining)

	default:
		return nil, fmt.Errorf("ParseControlPacket unknown %s", h)
	}
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

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_MQTT_Control_Packet
