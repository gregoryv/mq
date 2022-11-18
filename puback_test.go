package mq

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/eclipse/paho.golang/packets"
)

func ExamplePubAck_String() {
	p := NewPubAck()
	p.SetPacketID(9)
	fmt.Println(p)
	fmt.Print(DocumentFlags(p))
	// output:
	// PUBACK ---- p9 Success 4 bytes
	//        3210 PacketID Reason [reason text] Size
	//
	// 3-0 reserved
}

func ExampleNewPubRel() {
	p := NewPubRel()
	p.SetPacketID(9)
	p.SetReasonCode(PacketIdentifierNotFound)
	fmt.Println(p)
	// output:
	// PUBREL --1- p9 PacketIdentifierNotFound 5 bytes
}

func TestPubAck(t *testing.T) {
	p := NewPubAck()
	if v := p.String(); !strings.Contains(v, "PUBACK") {
		t.Error(v)
	}

	eq(t, p.SetPacketID, p.PacketID, 99)
	// should cover the check for remaining len
	testControlPacket(t, p)

	eq(t, p.SetReasonCode, p.ReasonCode, TopicNameInvalid)
	eq(t, p.SetReason, p.Reason, "name too long")

	p.AddUserProp("color", "red")

	testControlPacket(t, p)

	if v := p.String(); !strings.Contains(v, "name too long") {
		t.Error(v)
	}

	// variations
	{
		p := NewPubRel()
		testControlPacket(t, p)
		if v := p.String(); !strings.Contains(v, "PUBREL") {
			t.Error(v)
		}
	}
	{
		p := NewPubRec()
		if v := p.String(); !strings.Contains(v, "PUBREC") {
			t.Error(v)
		}
	}
	{
		p := NewPubComp()
		if v := p.String(); !strings.Contains(v, "PUBCOMP") {
			t.Error(v)
		}
	}

	// type
	if a, b := NewPubAck(), NewPubRel(); a.AckType() == b.AckType() {
		t.Error("PubAck byte same as PubRel byte")
	}
}

func BenchmarkPubAck(b *testing.B) {
	b.Run("our", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			p := NewPubAck()
			p.SetPacketID(99)
			p.AddUserProp("color", "red")
			p.WriteTo(&buf)
			ReadPacket(&buf)
		}
	})
	b.Run("their", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			p := packets.NewControlPacket(packets.PUBACK)
			c := p.Content.(*packets.Puback)
			c.PacketID = 99
			c.Properties = &packets.Properties{}
			c.Properties.User = append(
				c.Properties.User, packets.User{"color", "red"},
			)
			p.WriteTo(&buf)
			packets.ReadPacket(&buf)
		}
	})
}
