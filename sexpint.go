package mqtt

import (
	"encoding/binary"
	"time"
)

// 3.1.2.11.2 Session Expiry Interval
//
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901048
type SessionExpiryInterval uint32

func (s SessionExpiryInterval) MarshalBinary() ([]byte, error) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, uint32(s))
	return data, nil
}

func (s *SessionExpiryInterval) UnmarshalBinary(data []byte) error {
	*s = SessionExpiryInterval(binary.BigEndian.Uint32(data))
	return nil
}

func (s SessionExpiryInterval) String() string {
	return s.Duration().String()
}

func (s SessionExpiryInterval) Duration() time.Duration {
	return time.Duration(s) * time.Second
}
