package mqtt

// 2.1.1 Fixed Header
// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_MQTT_Control_Packet
type FixedHeader []byte

func (h FixedHeader) Name() string {
	return controlPacketTypeName[byte(h[0])&0b1111_0000]
}

func (h FixedHeader) Value() byte {
	return byte(h[0]) & 0b1111_0000
}

func (h FixedHeader) HasFlag(f byte) bool {
	return h.Flags()&f == f
}

func (h FixedHeader) Flags() byte {
	return byte(h[0]) & 0b0000_1111
}
