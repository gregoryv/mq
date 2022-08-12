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

	data := []byte{1}
	p.payload = &limitedReader{
		src:   bytes.NewReader(data),
		width: len(data),
	}

	var buf bytes.Buffer
	n, err := p.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if data, exp := buf.Bytes(), int64(11); n != exp {
		t.Log(data)
		t.Errorf("len(data) = %v, expected %v", n, exp)
	}

	if got, exp := p.String(), "CONNECT ----"; got != exp {
		t.Errorf("got %q, expected %q", got, exp)
	}
}

func BenchmarkControlPacket_Buffers(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := NewConnect()
		_, _ = p.WriteTo(ioutil.Discard)
	}
}

func TestSizeof(t *testing.T) {
	var p Connect
	_ = p
	best := uint(32)
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
