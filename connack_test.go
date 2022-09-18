package mqtt

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"
	"unsafe"
)

func ExampleConnAck() {
	a := NewConnAck()
	a.SetSessionPresent(true)

	fmt.Print(a.String())
	// output:
	// CONNACK ---- -------s  5 bytes
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

	if v := a.Flags(); v != 1 {
		t.Errorf("flags: %08b", v)
	}
	if !a.HasFlag(SessionPresent) {
		t.Error("HasFlag should be true for 1 if sessionPresent is set")
	}

	if false {
		var buf bytes.Buffer
		a.WriteTo(&buf)
		t.Logf("\n\n%s\n\n%s\n\n%v bytes", a.String(), hex.Dump(buf.Bytes()), size)
	}

	testControlPacket(t, &a)
}

func makeConnAck() ConnAck {
	a := NewConnAck()
	return a
}
