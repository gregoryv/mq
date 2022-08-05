package mqtt

import (
	"fmt"
	"io"
)

// 1.5.5 Variable Byte Integer
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901011
func ParseVarInt(r io.Reader) (uint, error) {
	var multiplier uint = 1
	var value uint
	buf := make([]byte, 1)
	var encodedByte byte
	for {
		if _, err := r.Read(buf); err != nil {
			return 0, err
		}
		encodedByte = buf[0]
		value += uint(encodedByte) & uint(127) * multiplier
		if multiplier > 128*128*128 {
			return 0, ErrMalformedVarInt
		}
		if encodedByte&128 == 0 {
			break
		}
		multiplier = multiplier * 128
	}
	return value, nil
}

var ErrMalformedVarInt = fmt.Errorf("malformed variable byte integer")

// 1.5.5 Variable Byte Integer
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901011
func NewVarInt(x uint) []byte {
	result := make([]byte, 0, 4) // max four
	if x == 0 {
		result = append(result, 0)
		return result
	}
	for x > 0 {
		encodedByte := byte(x % 128)
		x = x / 128
		if x > 0 {
			encodedByte = encodedByte | 128
		}
		result = append(result, encodedByte)
	}
	return result
}
