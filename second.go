package mqtt

import (
	"encoding/binary"
	"io"
	"time"
)

// connect properties

type Second uint16

func (s Second) WriteTo(w io.Writer) (int64, error) {
	data, err := s.MarshalBinary()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(data)
	return int64(n), err
}

func (s Second) MarshalBinary() ([]byte, error) {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, uint16(s))
	return data, nil
}

func (s *Second) UnmarshalBinary(data []byte) error {
	_ = data[1]
	*s = Second(binary.BigEndian.Uint16(data))
	return nil
}

func (s Second) Duration() time.Duration {
	return time.Duration(s) * time.Second
}
