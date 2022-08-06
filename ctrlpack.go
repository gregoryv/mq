package mqtt

import (
	"fmt"
)

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_MQTT_Control_Packet
type ControlPacket struct {
	FixedHeader
}

func (p *ControlPacket) String() string {
	return fmt.Sprintf("%s", p.FixedHeader.String())
}
