package mqtt

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"reflect"
	"testing"
)

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

	// break the variable length which should trigger an UnmarshalBinary
	bad := buf.Bytes()
	bad[3] = bad[3] + 1
	if _, err := ReadPacket(bytes.NewReader(bad)); err == nil {
		t.Error("expected error")
	}
}

// test helper for each control packet, should be called from each
// specific test e.g. TestPublish
func testControlPacket(in ControlPacket) error {
	// write it out
	var buf bytes.Buffer
	if _, err := in.WriteTo(&buf); err != nil {
		return err
	}
	data := make([]byte, buf.Len())
	copy(data, buf.Bytes())

	// read it back in
	got, err := ReadPacket(&buf)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(in, got) {
		got.WriteTo(&buf)
		return &diffErr{
			in:  fmt.Sprintf("%#v\n%s\n%s", in, in.String(), hex.Dump(data)),
			out: fmt.Sprintf("%#v\n%s\n%s", got, got.String(), hex.Dump(buf.Bytes())),
		}
	}
	return nil
}

type diffErr struct {
	in  string
	out string
}

func (e *diffErr) Error() string {
	return fmt.Sprintf("\n\nin\n%s\n\nout\n%s", e.in, e.out)
}
