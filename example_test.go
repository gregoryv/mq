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
		p.SetWill(mq.Pub(1, "client/gone", "pink"), 7)
		p.WriteTo(&buf)
	}
	{ // read the packet
		p, _ := mq.ReadPacket(&buf)
		fmt.Print(p)
	}
	// output:
	// CONNECT ---- up--1w-- MQTT5 pink 0s 58 bytes
}
