package mqtt

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func D(t *testing.T) {
	SetOutput(os.Stderr)
	t.Cleanup(func() {
		SetOutput(ioutil.Discard)
	})
}

func mustParse(t *testing.T, r io.Reader) interface{} {
	got, err := Parse(r)
	if err != nil {
		t.Helper()
		t.Fatal(err)
	}
	return got
}

func dump(v []byte) string {
	var buf bytes.Buffer
	for _, b := range v {
		fmt.Fprintf(&buf, "%08b: %q\n", b, b)
	}
	return buf.String()
}
