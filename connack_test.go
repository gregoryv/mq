package mqtt

import (
	"bytes"
	"encoding/hex"
	"testing"
	"unsafe"
)

func TestConnAck(t *testing.T) {
	a := NewConnAck()
	size := unsafe.Sizeof(*a)

	eq(t, a.SetSessionPresent, a.SessionPresent, true)

	eq(t, a.SetSessionExpiryInterval, a.SessionExpiryInterval, 199)
	eq(t, a.SetReceiveMax, a.ReceiveMax, 81)
	eq(t, a.SetMaxQoS, a.MaxQoS, 1)
	eq(t, a.SetRetainAvailable, a.RetainAvailable, true)
	eq(t, a.SetMaxPacketSize, a.MaxPacketSize, 250)
	eq(t, a.SetAssignedClientID, a.AssignedClientID, "macy")
	eq(t, a.SetTopicAliasMax, a.TopicAliasMax, 11)
	eq(t, a.SetReasonString, a.ReasonString, "because")

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
	var buf bytes.Buffer
	a.WriteTo(&buf)

	t.Logf("\n\n%s\n\n%s\n\n%v bytes", a, hex.Dump(buf.Bytes()), size)

}

var _ wireType = &ConnAck{}
