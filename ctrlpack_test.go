package mqtt

import (
	"fmt"
)

func ExampleControlPacket_String() {
	p := ControlPacket{
		FixedHeader: []byte{CONNECT, 0x00},
	}
	fmt.Println(p.String())
	// output:
	// CONNECT
}
