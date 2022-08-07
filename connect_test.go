package mqtt

import "fmt"

func ExampleConnect_String() {
	p := NewConnect().WithFlags(0b1111_1111)
	fmt.Println(p)
	fmt.Println(p.WithFlags(0b0000_0000))
	// output:
	// CONNECT 10 MQTT upr2wsR
	// CONNECT 10 MQTT -------
}
