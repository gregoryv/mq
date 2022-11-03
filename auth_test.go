package mq

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/eclipse/paho.golang/packets"
)

func ExampleAuth_String() {
	p := NewAuth()
	fmt.Println(p)
	fmt.Print(DocumentFlags(p))
	// output:
	// AUTH ---- 2 bytes
	//      3210 Size
	//
	// 3-0 reserved
}

func TestAuth(t *testing.T) {
	p := NewAuth()
	// normal disconnect
	testControlPacket(t, p)

	eq(t, p.SetReasonCode, p.ReasonCode, MalformedPacket)
	p.AddUserProp("color", "red")
	testControlPacket(t, p)

	// String
	if v := p.String(); v != "AUTH ---- 17 bytes" {
		t.Error(v)
	}
}

func BenchmarkAuth(b *testing.B) {
	b.Run("our", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			p := NewAuth()
			p.AddUserProp("color", "red")
			p.SetReasonCode(ReAuthenticate)
			p.WriteTo(&buf)
			ReadPacket(&buf)
		}
	})
	b.Run("their", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			p := packets.NewControlPacket(packets.AUTH)
			c := p.Content.(*packets.Auth)
			c.ReasonCode = packets.AuthReauthenticate
			c.Properties = &packets.Properties{}
			c.Properties.User = append(
				c.Properties.User, packets.User{"color", "red"},
			)
			p.WriteTo(&buf)
			packets.ReadPacket(&buf)
		}
	})
}
