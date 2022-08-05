package parser

import (
	"bytes"
	"context"
	"fmt"

	"github.com/gregoryv/mqtt"
)

func ExampleNewParser() {
	c := make(chan *mqtt.ControlPacket, 10)
	var con bytes.Buffer // some network connection
	parser := NewParser(&con)
	go parser.Parse(context.Background(), c)
	pak := <-c
	fmt.Println(pak)
	// output:
	// UNDEFINED
}
