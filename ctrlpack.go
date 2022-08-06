package mqtt

// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_MQTT_Control_Packet
type ControlPacket interface {
	FixedHeader() FixedHeader
}
