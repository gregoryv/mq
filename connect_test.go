package mqtt

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestConnect(t *testing.T) {
	p := NewConnect()
	p.clientID = "client"
	var buf bytes.Buffer
	p.WriteTo(&buf)
	t.Log(
		hex.Dump(buf.Bytes()),
		"\n", buf.Len(), "bytes",
	)
}

func TestconnectFlags(t *testing.T) {
	f := connectFlags(0b11110110)
	// QoS2
	if got, exp := f.String(), "upr2ws-"; got != exp {
		t.Errorf("got %q != exp %q", got, exp)
	}
	// QoS1
	f = connectFlags(0b11101110)
	if got, exp := f.String(), "upr1ws-"; got != exp {
		t.Errorf("got %q != exp %q", got, exp)
	}
	f = connectFlags(0b00000001)
	if got, exp := f.String(), "------!"; got != exp {
		t.Errorf("got %q != exp %q", got, exp)
	}
	if f.Has(WillFlag) || !f.Has(Reserved) {
		t.Errorf("Has %08b", f)
	}
}
