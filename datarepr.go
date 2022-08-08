package mqtt

import (
	"encoding/binary"
	"fmt"
)

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901007
type Bits byte

func (v Bits) Has(b Bits) bool {
	return v&b == b
}

// ----------------------------------------

type FourByteInt uint32

func (v FourByteInt) MarshalBinary() ([]byte, error) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, uint32(v))
	return data, nil
}

func (v *FourByteInt) UnmarshalBinary(data []byte) error {
	*v = FourByteInt(binary.BigEndian.Uint32(data))
	return nil
}

// ----------------------------------------

// 1.5.5 Variable Byte Integer
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901011
type VarInt uint

func (v VarInt) MarshalBinary() ([]byte, error) {
	data := make([]byte, 0, 4) // max four
	if v == 0 {
		data = append(data, 0)
		return data, nil
	}
	for v > 0 {
		encodedByte := byte(v % 128)
		v = v / 128
		if v > 0 {
			encodedByte = encodedByte | 128
		}
		data = append(data, encodedByte)
	}
	return data, nil
}

// ParseVarInt returns variable int from the reader. Returns EOF or
// wrapped ErrMalformed.
//
// 1.5.5 Variable Byte Integer
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901011
func (v *VarInt) UnmarshalBinary(data []byte) error {
	var multiplier uint = 1
	var value uint
	for _, encodedByte := range data {
		value += uint(encodedByte) & uint(127) * multiplier
		if multiplier > 128*128*128 {
			return fmt.Errorf("ParseVarInt: %w", ErrMalformed)
		}
		if encodedByte&128 == 0 {
			break
		}
		multiplier = multiplier * 128
	}
	*v = VarInt(value)
	return nil
}

var ErrMalformed = fmt.Errorf("malformed")
