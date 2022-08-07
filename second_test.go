package mqtt

import (
	"fmt"
)

func ExampleSecond() {
	a := Second(256)
	data, _ := a.MarshalBinary()

	var b Second
	b.UnmarshalBinary(data)

	fmt.Print(data, a.Duration(), b)

	// output:
	// [1 0] 4m16s 256
}
