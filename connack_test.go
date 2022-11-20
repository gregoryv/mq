package mq

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
	"unsafe"

	"github.com/eclipse/paho.golang/packets"
)

func ExampleConnAck() {
	a := NewConnAck()
	a.SetSessionPresent(true)
	a.SetReasonCode(NotAuthorized)

	fmt.Print(a.String())
	// output:
	// CONNACK ---- -------s  NotAuthorized 5 bytes
}

func ExampleConnAck_String() {
	a := NewConnAck()
	a.SetSessionPresent(true)
	a.SetAssignedClientID("pink")

	fmt.Println(a.String())
	fmt.Print(DocumentFlags(a))
	// output:
	// CONNACK ---- -------s pink 12 bytes
	//         3210 76543210 AssignedClientID Size
	//
	// 3-0 reserved
	//
	// 7-1 reserved
	// 0 s Session present
}

func TestConnAck(t *testing.T) {
	a := NewConnAck()
	size := unsafe.Sizeof(a)

	eq(t, a.SetSessionPresent, a.SessionPresent, true)
	eq(t, a.SetSessionExpiryInterval, a.SessionExpiryInterval, 199)
	eq(t, a.SetReceiveMax, a.ReceiveMax, 81)
	eq(t, a.SetMaxQoS, a.MaxQoS, 1)
	eq(t, a.SetRetainAvailable, a.RetainAvailable, true)
	eq(t, a.SetMaxPacketSize, a.MaxPacketSize, 250)
	eq(t, a.SetAssignedClientID, a.AssignedClientID, "macy")
	eq(t, a.SetTopicAliasMax, a.TopicAliasMax, 11)
	eq(t, a.SetReasonString, a.ReasonString, "because")
	eq(t, a.SetReasonCode, a.ReasonCode, UnspecifiedError)

	a.AddUserProp("color", "red")

	eq(t, a.SetWildcardSubAvailable, a.WildcardSubAvailable, true)
	eq(t, a.SetSubIdentifiersAvailable, a.SubIdentifiersAvailable, true)
	eq(t, a.SetSharedSubAvailable, a.SharedSubAvailable, true)
	eq(t, a.SetServerKeepAlive, a.ServerKeepAlive, 214)
	eq(t, a.SetResponseInformation, a.ResponseInformation, "gopher")
	eq(t, a.SetServerReference, a.ServerReference, "gopher")
	eq(t, a.SetAuthMethod, a.AuthMethod, "digest")
	eq(t, a.SetAuthData, a.AuthData, []byte("secret"))

	if !a.HasFlag(SessionPresent) {
		t.Error("HasFlag should be true for 1 if sessionPresent is set")
	}

	if false {
		var buf bytes.Buffer
		a.WriteTo(&buf)
		t.Logf("\n\n%s\n\n%s\n\n%v bytes", a.String(), hex.Dump(buf.Bytes()), size)
	}

	testControlPacket(t, a)
}

func BenchmarkConnAck(b *testing.B) {
	b.Run("our", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			p := NewConnAck()
			p.SetAuthMethod("digest")
			p.SetAuthData([]byte("secret"))
			p.SetSessionExpiryInterval(30)
			p.AddUserProp("color", "red")
			p.WriteTo(&buf)
			ReadPacket(&buf)
		}
	})
	b.Run("their", func(b *testing.B) {
		var buf bytes.Buffer
		for i := 0; i < b.N; i++ {
			p := packets.NewControlPacket(packets.CONNACK)
			c := p.Content.(*packets.Connack)
			c.Properties = &packets.Properties{}
			c.Properties.AuthMethod = "digest"
			c.Properties.AuthData = []byte("secret")
			sExpiry := uint32(30)
			c.Properties.SessionExpiryInterval = &sExpiry
			c.Properties.User = append(
				c.Properties.User, packets.User{"color", "red"},
			)
			p.WriteTo(&buf)
			packets.ReadPacket(&buf)
		}
	})
}
