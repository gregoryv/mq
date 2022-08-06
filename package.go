package mqtt

import (
	"github.com/gregoryv/mqtt/internal"
)

// Fixed header flags
const (
	RETAIN byte = 0b0000_0001
	QoS0   byte = 0b0000_0000
	QoS1   byte = 0b0000_0010
	QoS2   byte = 0b0000_0100
	//QoS3 FixedHeader = 0b0000_0110   malformed!
	DUP byte = 0b0000_1000
)

var FlagNames = internal.NewDict(
	map[byte]string{
		DUP:    "DUP",
		QoS0:   "QoS0",
		QoS1:   "QoS1",
		QoS2:   "QoS2",
		RETAIN: "RETAIN",
	},
)
