/*
Package mqtt provides a MQTT v5.0 protocol implementation

The specification is found at
https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html

*/
package mqtt

import (
	"encoding"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// wireType defines the interface for types that can be send over the
// wire
type wireType interface {
	encoding.BinaryUnmarshaler

	// fill unmarshals the data type into buf at position i. The
	// returned value is the width of the data marshaled.  fill should
	// work with a nil buf as a noop but return the width.  This
	// enables efficient calculation of partial lengths without
	// actually allocating a buf.
	fill(buf []byte, i int) int

	// fillProp fills the identified property if not empty as this is
	// the case for most property values.
	fillProp(buf []byte, i int, id Ident) int

	// returns the width of the wire data in bytes
	width() int
}

// firstByte represents the first byte in a control packet.
type firstByte byte

// String returns a readable string TYPEFLAGS, e.g. PUBLISH d1-r
func (f firstByte) String() string {
	var sb strings.Builder
	sb.WriteString(typeNames[byte(f)&0b1111_0000])
	sb.WriteString(" ")
	flags := []byte("----")
	if Bits(f).Has(DUP) {
		flags[0] = 'd'
	}
	switch {
	case Bits(f).Has(QoS3):
		flags[1] = '!' // malformed
		flags[2] = '!' // malformed
	case Bits(f).Has(QoS1):
		flags[2] = '1'
	case Bits(f).Has(QoS2):
		flags[1] = '2'
	}
	if Bits(f).Has(RETAIN) {
		flags[3] = 'r'
	}
	sb.Write(flags)
	return sb.String()
}

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901013
type property [2]string

func (v property) fillProp(data []byte, i int, id Ident) int {
	if len(v[0]) == 0 {
		return 0
	}
	n := i
	i += id.fill(data, i)
	i += v.fill(data, i)
	return i - n
}
func (v property) fill(data []byte, i int) int {
	i += wstring(v[0]).fill(data, i)
	_ = wstring(v[1]).fill(data, i)
	return v.width()
}

func (v *property) UnmarshalBinary(data []byte) error {
	var key wstring
	if err := key.UnmarshalBinary(data); err != nil {
		return unmarshalErr(v, "key", err.(*Malformed))
	}
	v[0] = string(key)

	i := len(v[0]) + 2
	var val wstring
	if err := val.UnmarshalBinary(data[i:]); err != nil {
		return unmarshalErr(v, "value", err.(*Malformed))
	}
	v[1] = string(val)
	return nil
}
func (v property) String() string {
	return fmt.Sprintf("%s:%s", v[0], v[1])
}
func (v property) width() int {
	return wstring(v[0]).width() + wstring(v[1]).width()
}

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901010
type wstring = bindata

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901012
type bindata []byte

func (v bindata) fillProp(data []byte, i int, id Ident) int {
	if len(v) == 0 {
		return 0
	}
	n := i
	i += id.fill(data, i)
	i += v.fill(data, i)
	return i - n
}
func (v bindata) fill(data []byte, i int) int {
	if len(data) >= i+v.width() {
		i += wuint16(len(v)).fill(data, i)
		copy(data[i:], []byte(v))
	}
	return v.width()
}

func (v *bindata) UnmarshalBinary(data []byte) error {
	var length wuint16
	_ = length.UnmarshalBinary(data)
	if len(data) < int(length)+2 {
		return unmarshalErr(v, "", "missing data")
	}
	if length == 0 {
		return nil
	}
	*v = make([]byte, length)
	copy(*v, data[2:length+2])
	return nil
}

func (v bindata) width() int {
	return 2 + len(v)
}

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901011
type vbint uint

func (v vbint) fillProp(data []byte, i int, id Ident) int {
	if v == 0 {
		return 0
	}
	n := i
	i += id.fill(data, i)
	i += v.fill(data, i)
	return i - n
}
func (v vbint) fill(data []byte, i int) int {
	x := v
	n := i
	for {
		encodedByte := byte(x % 128)
		x = x / 128
		if x > 0 {
			encodedByte = encodedByte | 128
		}
		if i < len(data) {
			data[i] = encodedByte
		}
		i++
		if x == 0 {
			break
		}
	}
	return i - n
}

func (v vbint) width() int {
	return v.fill(_LEN, 0)
}

func (v *vbint) ReadFrom(r io.Reader) (int64, error) {
	var multiplier uint = 1
	var value uint
	data := make([]byte, 1)
	var i int64
	for {
		if _, err := r.Read(data); err != nil {
			return i, err
		}
		i++
		encodedByte := data[0]
		value += uint(encodedByte) & uint(127) * multiplier
		if multiplier > 128*128*128 {
			return i, unmarshalErr(v, "", "size exceeded")
		}
		if encodedByte&128 == 0 {
			break
		}
		multiplier = multiplier * 128
	}
	*v = vbint(value)
	return i, nil
}

// UnmarshalBinary data, returns nil or *Malformed error
func (v *vbint) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return unmarshalErr(v, "", "missing data")
	}
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
	*v = vbint(value)
	return nil
}

// wire types
type (
	wuint8 = Bits // byte
)

type wbool bool

func (v wbool) fillProp(data []byte, i int, id Ident) int {
	if !v {
		return 0
	}
	n := i
	i += id.fill(data, i)
	i += v.fill(data, i)
	return i - n
}
func (v wbool) fill(data []byte, i int) int {
	if len(data) >= i+1 {
		if v {
			data[i] = 0x01
		} else {
			data[i] = 0x00
		}
	}
	return 1
}
func (v *wbool) UnmarshalBinary(data []byte) error {
	switch data[0] {
	case 0:
		*v = wbool(false)
	case 1:
		*v = wbool(true)
	default:
		return fmt.Errorf("malformed bool")
	}
	return nil
}
func (v wbool) width() int { return 1 }

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901007
type Bits byte

func (v Bits) Has(b byte) bool { return byte(v)&b == b }

func (v Bits) fillProp(data []byte, i int, id Ident) int {
	if v == 0 {
		return 0
	}
	n := i
	i += id.fill(data, i)
	i += v.fill(data, i)
	return i - n
}

func (v Bits) fill(data []byte, i int) int {
	if len(data) >= i+1 {
		data[i] = byte(v)
	}
	return 1
}

// fillOpt fills the bits if > 0
func (v Bits) fillOpt(data []byte, i int) int {
	if v == 0 {
		return 0
	}
	return v.fill(data, i)
}

func (v *Bits) ReadFrom(r io.Reader) (int64, error) {
	data := make([]byte, 1)
	if n, err := r.Read(data); err != nil {
		return int64(n), err
	}
	return 1, v.UnmarshalBinary(data)
}
func (v *Bits) UnmarshalBinary(data []byte) error {
	*v = Bits(data[0])
	return nil
}
func (v Bits) width() int { return 1 }
func (v *Bits) toggle(flag byte, on bool) {
	if on {
		*v = *v | Bits(flag)
		return
	}
	*v = *v & Bits(^flag)
}

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901008
type wuint16 uint16

func (v wuint16) fillProp(data []byte, i int, id Ident) int {
	if v == 0 {
		return 0
	}
	n := i
	i += id.fill(data, i)
	i += v.fill(data, i)
	return i - n
}

func (v wuint16) fill(data []byte, i int) int {
	if len(data) >= i+2 {
		binary.BigEndian.PutUint16(data[i:], uint16(v))
	}
	return 2
}

func (v *wuint16) UnmarshalBinary(data []byte) error {
	*v = wuint16(binary.BigEndian.Uint16(data))
	return nil
}

func (v wuint16) width() int { return 2 }

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901009
type wuint32 uint32

func (v wuint32) fillProp(data []byte, i int, id Ident) int {
	if v == 0 {
		return 0
	}
	n := i
	i += id.fill(data, i)
	i += v.fill(data, i)
	return i - n
}

func (v wuint32) fill(data []byte, i int) int {
	if len(data) >= i+v.width() {
		binary.BigEndian.PutUint32(data[i:], uint32(v))
	}
	return v.width()
}

func (v *wuint32) UnmarshalBinary(data []byte) error {
	*v = wuint32(binary.BigEndian.Uint32(data))
	return nil
}

func (v wuint32) width() int { return 4 }

// only here to fulfill interface
func (v Ident) fillProp(data []byte, i int, id Ident) int { return 0 }

func (v Ident) fill(data []byte, i int) int {
	if len(data) >= i+1 {
		data[i] = byte(v)
	}
	return 1
}

func (v *Ident) UnmarshalBinary(data []byte) error {
	*v = Ident(data[0])
	return nil
}

func (v Ident) width() int { return 1 }
