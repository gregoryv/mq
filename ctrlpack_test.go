package mqtt

import (
	"fmt"
)

func ExampleControlPacket() {
	p := &ControlPacket{
		data:           []byte{0x10, 0x00},
		endFixedHeader: 1,
	}
	fmt.Println(p)
	// output:
	// CONNECT
}
