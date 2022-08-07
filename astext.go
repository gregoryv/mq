package mqtt

import (
	"bytes"
	"encoding"
)

type AsText []encoding.TextMarshaler

func (t AsText) MarshalText() ([]byte, error) {
	var buf bytes.Buffer
	var lastErr error
	for _, v := range t {
		text, err := v.MarshalText()
		buf.Write(text)
		if err != nil {
			buf.WriteString(" ")
			buf.WriteString(err.Error())
			buf.WriteString("\n")
			lastErr = err
		}
	}
	return buf.Bytes(), lastErr
}
