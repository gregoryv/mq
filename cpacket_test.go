package mqtt

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/gregoryv/asserter"
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

	// Check the panic for now
	t.Run("done", func(t *testing.T) {
		defer func() {
			if err := recover(); err == nil {
				t.Error("time to remove me? are you done")
			}
		}()
		for i := 0; i < 16; i++ {
			data := []byte{byte(i) << 4, 3, 0, 0, 0}
			ReadPacket(bytes.NewReader(data))
		}
	})

}

// test helper for each control packet, should be called from each
// specific test e.g. TestPublish
func testControlPacket(t *testing.T, in ControlPacket) {
	// write it out
	var buf bytes.Buffer
	if _, err := in.WriteTo(&buf); err != nil {
		t.Error("WriteTo", err)
	}
	data := make([]byte, buf.Len())
	copy(data, buf.Bytes())

	a := strings.ReplaceAll(fmt.Sprintf("%#v", in), ", ", ",\n")
	t.Log(buf.Len(), "\n\n", a, "\n\n", hex.Dump(data))
	// read it back in
	got, err := ReadPacket(&buf)
	if err != nil {
		t.Fatal("ReadPacket", err)
	}

	if !reflect.DeepEqual(in, got) {
		got.WriteTo(&buf)

		b := strings.ReplaceAll(fmt.Sprintf("%#v", got), ", ", ",\n")

		assert := asserter.New(t)
		assert().Equals(b, a)
	}
}
