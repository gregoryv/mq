package mq

import "testing"

func TestPingReq(t *testing.T) {
	p := NewPingReq()

	testControlPacket(t, &p)

	if err := p.UnmarshalBinary(nil); err != nil {
		t.Error("PingReq.UnmarshalBinary should be a noop")
	}
}
