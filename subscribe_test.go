package mqtt

import "testing"

func TestSubscribe(t *testing.T) {
	s := NewSubscribe()

	eq(t, s.SetPacketID, s.PacketID, 34)
	eq(t, s.SetSubscriptionID, s.SubscriptionID, 99)

	s.AddUserProp("color", "purple")

	s.AddFilter("a/b/c", FopQoS2|FopNL|FopRAP) // todo define FilterOptions
	t.Log(&s)

	if err := testControlPacket(&s); err != nil {
		t.Error(err)
	}
}
