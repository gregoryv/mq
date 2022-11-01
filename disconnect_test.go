package mq

import (
	"bytes"
	"fmt"
	"testing"
)

func ExampleDisconnect_String() {
	p := NewDisconnect()
	fmt.Println(p)
	fmt.Print(DocumentFlags(p))
	// output:
	// DISCONNECT ---- 2 bytes
	//            3210 Size
	//
	// 3-0 reserved
}

func TestDisconnect(t *testing.T) {
	p := NewDisconnect()
	// normal disconnect
	testControlPacket(t, p)

	eq(t, p.SetReasonCode, p.ReasonCode, MalformedPacket)
	p.AddUserProp("color", "red")
	testControlPacket(t, p)
}

func BenchmarkDisconnect_UnmarshalBinary(b *testing.B) {
	p := NewDisconnect()
	var buf bytes.Buffer
	p.WriteTo(&buf)
	data := buf.Bytes()[1:] // without the fixed header

	var in Disconnect
	for i := 0; i < b.N; i++ {
		in.UnmarshalBinary(data[1:])
	}
}
