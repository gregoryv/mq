package mqtt

import "testing"

func TestAuth(t *testing.T) {
	p := NewAuth()
	// normal disconnect
	testControlPacket(t, &p)

	eq(t, p.SetReasonCode, p.ReasonCode, MalformedPacket)
	p.AddUserProp("color", "red")
	testControlPacket(t, &p)
}