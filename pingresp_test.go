package mq

import "testing"

func TestPingResp(t *testing.T) {
	p := NewPingResp()

	testControlPacket(t, &p)
	if err := p.UnmarshalBinary(nil); err != nil {
		t.Error("PingResp.UnmarshalBinary should be a noop")
	}
}
