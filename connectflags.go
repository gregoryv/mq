package mqtt

import "bytes"

// 3.1.2.3 Connect Flags
const (
	Reserved byte = 1 << iota
	CleanStart
	WillFlag
	WillQoS1
	WillQoS2
	WillRetain
	PasswordFlag
	UsernameFlag
)

type ConnectFlags byte

func (c ConnectFlags) String() string {
	flags := bytes.Repeat([]byte("-"), 7)
	for i, f := range connectFlagOrder {
		if c.Has(f) {
			flags[i] = shortConnectFlags[f]
		}
	}
	if c.Has(WillQoS1) {
		flags[3] = '1'
	}
	if c.Has(WillQoS2) {
		flags[3] = '2'
	}
	if c.Has(Reserved) {
		flags[6] = '!'
	}
	return string(flags)
}

func (c ConnectFlags) Has(f byte) bool { return byte(c)&f == f }

var shortConnectFlags = map[byte]byte{
	//	Reserved:     '',
	CleanStart:   's',
	WillFlag:     'w',
	WillQoS1:     '1',
	WillQoS2:     '2',
	WillRetain:   'r',
	PasswordFlag: 'p',
	UsernameFlag: 'u',
}

var connectFlagOrder = []byte{
	UsernameFlag, // bit 7
	PasswordFlag, // bit 6
	WillRetain,   // bit 5
	'-',          // QoS bits 4 and 3
	WillFlag,     // bit 2
	CleanStart,   // bit 1
	Reserved,     // bit 0
}
