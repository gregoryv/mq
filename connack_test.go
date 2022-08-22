package mqtt

import "testing"

func TestConnAck(t *testing.T) {
	a := NewConnAck()

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

	t.Logf("\n\n%s\n\n", a)
}

var _ wireType = &ConnAck{}
