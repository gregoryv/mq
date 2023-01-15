package mq

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/gregoryv/asserter"
)

func TestDump(t *testing.T) {
	packets := []Packet{
		NewAuth(),
		NewConnAck(),
		NewConnect(),
		NewDisconnect(),
		// NewPingReq(), empty by definition
		// NewPingResp(), empty by definition
		NewPubAck(),
		NewPubComp(),
		NewPublish(),
		NewPubRec(),
		NewPubRel(),
		NewSubAck(),
		NewSubscribe(),
		NewUnsubAck(),
		NewUnsubscribe(),
	}
	for _, p := range packets {
		var buf bytes.Buffer
		Dump(&buf, p)
		if buf.Len() == 0 {
			t.Errorf("%T is empty", p)
		}
	}
}

func TestReadPacket_broken(t *testing.T) {
	var r brokenRW
	if _, err := ReadPacket(&r); err == nil {
		t.Error("expected error")
	}

	var buf bytes.Buffer
	p := NewPublish()
	p.SetTopicName("a/b/c")
	p.WriteTo(&buf)

	partial := buf.Bytes()[:2]
	if _, err := ReadPacket(bytes.NewReader(partial)); err == nil {
		t.Error("expected error")
	}

	buf.Reset()
	p.WriteTo(&buf)

	// break the variable length which should trigger an error in
	// UnmarshalBinary
	bad := buf.Bytes()
	bad[3] = bad[3] - 1
	if _, err := ReadPacket(bytes.NewReader(bad)); err == nil {
		t.Error("expected error")
	}

	data := []byte{0, 3, 0, 0, 0}
	if v, err := ReadPacket(bytes.NewReader(data)); err != nil {
		t.Error("undefined should not fail", err, v)
	} else {
		v := v.(*Undefined)
		if len(v.Data()) != 3 {
			t.Error(v.Data())
		}
	}
}

// test helper for each control packet, should be called from each
// specific test e.g. TestPublish
func testControlPacket(t *testing.T, in ControlPacket) {
	//t.Helper()
	// write it out
	var buf bytes.Buffer
	if _, err := in.WriteTo(&buf); err != nil {
		t.Log(in)
		t.Error("WriteTo", err)
	}
	data := make([]byte, buf.Len())
	copy(data, buf.Bytes())

	a := strings.ReplaceAll(fmt.Sprintf("%#v", in), ", ", ",\n")

	// read it back in
	got, err := ReadPacket(&buf)
	if err != nil {
		t.Log(buf.Len(), "\n\n", a, "\n\n", hex.Dump(data))
		t.Fatal("ReadPacket", err)
	}

	if !reflect.DeepEqual(in, got) {
		var buf bytes.Buffer
		got.WriteTo(&buf)

		b := strings.ReplaceAll(fmt.Sprintf("%#v", got), ", ", ",\n")

		assert := asserter.New(t)
		assert().Equals(b, a)
		t.Log(len(data), "bytes\n\n", hex.Dump(data))
		t.Log(buf.Len(), "bytes\n\n", hex.Dump(buf.Bytes()))
	}

	// String
	if v := got.String(); !strings.Contains(v, " bytes") {
		t.Error("empty .String")
	}
}
