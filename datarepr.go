package mqtt

import (
	"encoding/binary"
	"fmt"
)

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901013
type UTF8StringPair [2]UTF8String

func (v UTF8StringPair) MarshalBinary() ([]byte, error) {
	key, err := v[0].MarshalBinary()
	if err != nil {
		return nil, marshalErr(v, "key", err.(*Malformed))
	}
	val, err := v[1].MarshalBinary()
	if err != nil {
		return nil, marshalErr(v, "value", err.(*Malformed))
	}
	return append(key, val...), nil
}

func (v *UTF8StringPair) UnmarshalBinary(data []byte) error {
	if err := v[0].UnmarshalBinary(data); err != nil {
		return unmarshalErr(v, "key", err.(*Malformed))
	}
	i := len(v[0]) + 2
	if err := v[1].UnmarshalBinary(data[i:]); err != nil {
		return unmarshalErr(v, "value", err.(*Malformed))
	}
	return nil
}
func (v UTF8StringPair) String() string {
	return fmt.Sprintf("%s:%s", v[0], v[1])
}

// ----------------------------------------

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901010
type UTF8String string

func (v UTF8String) MarshalBinary() ([]byte, error) {
	data, err := BinaryData([]byte(v)).MarshalBinary()
	if err != nil {
		return data, marshalErr(v, "", err.(*Malformed))
	}
	return data, nil
}

func (v *UTF8String) UnmarshalBinary(data []byte) error {
	var b BinaryData
	if err := b.UnmarshalBinary(data); err != nil {
		return unmarshalErr(v, "", err.(*Malformed))
	}
	*v = UTF8String(b)
	return nil
}

// ----------------------------------------

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901012
type BinaryData []byte

func (v BinaryData) MarshalBinary() ([]byte, error) {
	if len(v) > MaxUint16 {
		return nil, marshalErr(v, "", "size exceeded")
	}
	data := make([]byte, len(v)+2)
	l, _ := TwoByteInt(len(v)).MarshalBinary()
	copy(data[:2], l)
	copy(data[2:], []byte(v))
	return data, nil
}

func (v *BinaryData) UnmarshalBinary(data []byte) error {
	var l TwoByteInt
	_ = l.UnmarshalBinary(data)
	if len(data) < int(l)+2 {
		return unmarshalErr(v, "", "missing data")
	}
	*v = make([]byte, l)
	copy(*v, data[2:l+2])
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

// UnmarshalBinary data, returns nil or *Malformed error
func (v *VarByteInt) UnmarshalBinary(data []byte) error {
	var multiplier uint = 1
	var value uint
	for _, encodedByte := range data {
		value += uint(encodedByte) & uint(127) * multiplier
		if multiplier > 128*128*128 {
			return unmarshalErr(v, "", "size exceeded")
		}
		if encodedByte&128 == 0 {
			break
		}
		multiplier = multiplier * 128
	}
	*v = VarByteInt(value)
	return nil
}

func (v VarByteInt) Width() int {
	switch {
	case v < 128:
		return 1
	case v < 16_384:
		return 2
	case v < 2_097_152:
		return 3
	default:
		return 4
	}
}

// ----------------------------------------
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901007
type Bits byte

func (v Bits) Has(b byte) bool {
	return byte(v)&b == b
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

func marshalErr(v interface{}, ref string, err interface{}) *Malformed {
	e := newMalformed(v, ref, err)
	e.method = "marshal"
	return e
}

func unmarshalErr(v interface{}, ref string, err interface{}) *Malformed {
	e := newMalformed(v, ref, err)
	e.method = "unmarshal"
	return e
}

func newMalformed(v interface{}, ref string, err interface{}) *Malformed {
	var reason string
	switch e := err.(type) {
	case *Malformed:
		reason = e.reason
	case string:
		reason = e
	}
	// remove * from type name
	t := fmt.Sprintf("%T", v)
	if t[0] == '*' {
		t = t[1:]
	}
	return &Malformed{
		t:      t,
		ref:    ref,
		reason: reason,
	}
}

type Malformed struct {
	method string
	t      string
	ref    string
	reason string
}

func (e *Malformed) Error() string {
	if e.ref == "" {
		return fmt.Sprintf("malformed %s %s: %s", e.t, e.method, e.reason)
	}
	return fmt.Sprintf("malformed %s %s: %s %s", e.t, e.method, e.ref, e.reason)
}

// see math.MaxUint16
const MaxUint16 = 1<<16 - 1
