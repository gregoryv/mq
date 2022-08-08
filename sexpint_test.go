package mqtt

import "testing"

func TestSessionExpiryInterval(t *testing.T) {
	var s SessionExpiryInterval

	if _, err := s.MarshalBinary(); err != nil {
		t.Error("MarshalBinary", err)
	}
}
