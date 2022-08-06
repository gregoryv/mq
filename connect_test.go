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
	// 00000000: '\x00'
	// 00000100: '\x04'
	// 01001101: 'M'
	// 01010001: 'Q'
	// 01010100: 'T'
	// 01010100: 'T'
}
