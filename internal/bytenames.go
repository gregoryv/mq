package internal

import "strings"

func NewByteNames(m map[byte]string) *ByteNames {
	return &ByteNames{names: m}
}

type ByteNames struct {
	names map[byte]string
}

func (n *ByteNames) Name(b byte) string {
	return n.names[b]
}

func (n *ByteNames) Join(sep string, b []byte) string {
	if len(b) == 0 {
		return ""
	}
	names := make([]string, len(b), len(b))
	for i, b := range b {
		names[i] = n.Name(b)
	}
	return strings.Join(names, sep)
}
