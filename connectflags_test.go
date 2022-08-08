package mqtt

import "testing"

func TestConnectFlags(t *testing.T) {
	f := ConnectFlags(0b11110110)
	// QoS2
	if got, exp := f.String(), "upr2ws-"; got != exp {
		t.Errorf("got %q != exp %q", got, exp)
	}
	// QoS1
	f = ConnectFlags(0b11101110)
	if got, exp := f.String(), "upr1ws-"; got != exp {
		t.Errorf("got %q != exp %q", got, exp)
	}
	f = ConnectFlags(0b00000001)
	if got, exp := f.String(), "------R"; got != exp {
		t.Errorf("got %q != exp %q", got, exp)
	}
	if f.Has(WillFlag) || !f.Has(Reserved) {
		t.Errorf("Has %08b", f)
	}
}
