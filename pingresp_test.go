package mqtt

import "testing"

func TestPingResp(t *testing.T) {
	p := NewPingResp()

	testControlPacket(t, &p)
}
