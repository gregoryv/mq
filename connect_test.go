package mqtt

import (
	"bytes"
	"io/ioutil"
	"testing"
	"unsafe"
)

func TestConnect(t *testing.T) {
	p := NewConnect()

	if err := Check(p); err == nil {
		t.Error("should fail to write an empty connect")
	}

	p.SetWillDelayInterval(10)

	var buf bytes.Buffer
	if _, err := p.WriteTo(&buf); err != nil {
		t.Fatal(err)
	}

	// once closer to final implementation do a byte check

	if got, exp := p.String(), "CONNECT ---- ----w--"; got != exp {
		t.Errorf("got %q, expected %q", got, exp)
	}
}

func BenchmarkControlPacket_Buffers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewConnect()
		_, _ = p.WriteTo(ioutil.Discard)
	}
}

// todo reactivate once closing in with implementation
func xTestSizeof(t *testing.T) {
	var p Connect
	best := 56
	got := int(unsafe.Sizeof(p))
	switch {
	case got > best:
		t.Errorf("ControlPacket size increased from %v to: %v", best, got)
	case got < best:
		t.Errorf(
			"Packet size improved from %v to %v, update TestSizeof",
			best, got,
		)
	}
}

// ---------------------------------------------------------------------
// 3.1.2.3 Connect Flags
// ---------------------------------------------------------------------

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
	if got, exp := f.String(), "------!"; got != exp {
		t.Errorf("got %q != exp %q", got, exp)
	}
	if f.Has(WillFlag) || !f.Has(Reserved) {
		t.Errorf("Has %08b", f)
	}
}
