package mqtt

import (
	"fmt"
)

func ExampleNewConnect() {
	cp := NewConnect()
	fmt.Println(cp.FixedHeader())
	fmt.Println(dump(cp.Bytes()))
	// output:
	// CONNECT 6
	// 00010000: '\x10'
	// 00000110: '\x06'
}
