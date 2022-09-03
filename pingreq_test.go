package mqtt

import "testing"

func TestPingReq(t *testing.T) {
	p := NewPingReq()

	testControlPacket(t, &p)
}
