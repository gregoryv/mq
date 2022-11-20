package mq

import "fmt"

func ExampleMalformed_Error() {
	var e Malformed
	e.SetPacket(NewConnect())
	e.SetReason("missing data")
	fmt.Println(e.Error())
	// output:
	// malformed *mq.Connect: missing data
}
