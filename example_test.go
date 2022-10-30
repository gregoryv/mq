package mq_test

import (
	"bytes"
	"fmt"

	"github.com/gregoryv/mq"
)

func Example_packetReadWrite() {
	var buf bytes.Buffer
	{ // create and write packet
		p := mq.NewConnect()
		p.SetClientID("pink")
		p.SetUsername("gopher")
		p.SetPassword([]byte("cute"))
		p.SetWillQoS(1)
		p.WriteTo(&buf)
	}
	{ // read the packet
		p, _ := mq.ReadPacket(&buf)
		fmt.Print(p)
	}
	// output:
	// CONNECT ---- up--1--- MQTT5 pink 0s 33 bytes
}
