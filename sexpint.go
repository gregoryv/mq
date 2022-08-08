package mqtt

import (
	"encoding/binary"
	"time"
)

type SessionExpiryInterval uint32

func (s SessionExpiryInterval) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s SessionExpiryInterval) String() string {
	return s.Duration().String()
}
func (s SessionExpiryInterval) MarshalBinary() ([]byte, error) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, uint32(s))
	return data, nil
}

func (s *SessionExpiryInterval) UnmarshalBinary(data []byte) error {
	*s = SessionExpiryInterval(binary.BigEndian.Uint32(data))
	return nil
}

func (s SessionExpiryInterval) Duration() time.Duration {
	return time.Duration(s) * time.Second
}
