package mqtt

import (
	"bytes"
	"fmt"
)

func dump(v []byte) string {
	var buf bytes.Buffer
	for _, b := range v {
		fmt.Fprintf(&buf, "%08b: %q\n", b, b)
	}
	return buf.String()
}
