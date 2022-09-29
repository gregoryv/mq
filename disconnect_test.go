package mq

import "testing"

func TestDisconnect(t *testing.T) {
	p := NewDisconnect()
	// normal disconnect
	testControlPacket(t, &p)

	eq(t, p.SetReasonCode, p.ReasonCode, MalformedPacket)
	p.AddUserProp("color", "red")
	testControlPacket(t, &p)
}
