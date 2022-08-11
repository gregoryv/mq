package mqtt

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
	"time"
	"unsafe"
)

func TestConnect_Buffers(t *testing.T) {
	t.Fatal("prove the correctness of CONNECT packet Buffers")
	p := NewConnect()
	bin, err := p.Buffers()
	if err != nil {
		var buf bytes.Buffer
		bin.WriteTo(&buf)
		t.Error(hex.Dump(buf.Bytes()), "\n", err)
	}
}

func TestConnect_String(t *testing.T) {
	cases := map[string]*Connect{
		"CONNECT": NewConnect(),
	}
	for exp, p := range cases {
		if got := p.String(); !strings.HasPrefix(got, exp) {
			t.Error(got)
		}
	}
}

func BenchmarkControlPacket_Buffers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewConnect()
		_, _ = p.Buffers()
	}
}

func TestSizeof(t *testing.T) {
	var p Connect
	_ = p
	best := uint(48)
	got := uint(unsafe.Sizeof(p))
	switch {
	case got > best:
		t.Error("ControlPacket size: ", got)
	case got < best:
		t.Errorf("Size %v improved, update TestSizeof", got)
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

// ---------------------------------------------------------------------
// 3.1.2.11 CONNECT Properties
// ---------------------------------------------------------------------

func TestConnectProperties(t *testing.T) {
	b := SessionExpiryInterval(76)

	if got := b.String(); got != "1m16s" {
		t.Error("unexpected text", got)
	}

	if dur := b.Duration(); dur != 76*time.Second {
		t.Error("unexpected duration", dur)
	}
}
