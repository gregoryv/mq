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

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901008
type TwoByteInt uint16

func (v TwoByteInt) MarshalBinary() ([]byte, error) {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, uint16(v))
	return data, nil
}

func (v *TwoByteInt) UnmarshalBinary(data []byte) error {
	*v = TwoByteInt(binary.BigEndian.Uint16(data))
	return nil
}

// ----------------------------------------

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901009
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

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901010
type UTF8String string

func (v UTF8String) MarshalBinary() ([]byte, error) {
	if len(v) > MaxUint16 {
		return nil, ErrUTF8StringTooLarge
	}
	data := make([]byte, len(v)+2)
	l, _ := TwoByteInt(len(v)).MarshalBinary()
	copy(data[:2], l)
	copy(data[2:], []byte(v))
	return data, nil
}

func (v *UTF8String) UnmarshalBinary(data []byte) error {
	var l TwoByteInt
	_ = l.UnmarshalBinary(data)
	if int(l) != len(data)-2 {
		return ErrMissingData
	}
	*v = UTF8String(data[2 : l+2])
	return nil
}

// ----------------------------------------

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901011
type VarByteInt uint

func (v VarByteInt) MarshalBinary() ([]byte, error) {
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

// UnmarshalBinary returns nil or fails with EOF or ErrMalformed.
func (v *VarByteInt) UnmarshalBinary(data []byte) error {
	var multiplier uint = 1
	var value uint
	for _, encodedByte := range data {
		value += uint(encodedByte) & uint(127) * multiplier
		if multiplier > 128*128*128 {
			return ErrMalformed
		}
		if encodedByte&128 == 0 {
			break
		}
		multiplier = multiplier * 128
	}
	*v = VarByteInt(value)
	return nil
}

// ----------------------------------------

var (
	ErrMalformed          = fmt.Errorf("malformed")
	ErrMissingData        = fmt.Errorf("missing data")
	ErrUTF8StringTooLarge = fmt.Errorf("utf8 string too large")
)

// see math.MaxUint16
const MaxUint16 = 1<<16 - 1
