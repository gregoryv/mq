package mqtt

import (
	"encoding/binary"
)

// 3.1.2.11.3 Receive Maximum
//
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901049
type ReceiveMax uint16

func (r ReceiveMax) MarshalBinary() ([]byte, error) {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, uint16(r))
	return data, nil
}

func (r *ReceiveMax) UnmarshalBinary(data []byte) error {
	*r = ReceiveMax(binary.BigEndian.Uint16(data))
	return nil
}
