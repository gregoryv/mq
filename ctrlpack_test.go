package mqtt

import (
	"fmt"
)

func ExampleControlPacket_String() {
	p := ControlPacket{
		FixedHeader: []byte{CONNECT, 2},
	}
	fmt.Println(p.String())
	// output:
	// CONNECT 2
}
