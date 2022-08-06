package internal

import "strings"

func NewDict(m map[byte]string) *Dict {
	return &Dict{names: m}
}

type Dict struct {
	names map[byte]string
}

func (n *Dict) Name(b byte) string {
	return n.names[b]
}

func (n *Dict) Join(sep string, b []byte) string {
	if len(b) == 0 {
		return ""
	}
	names := make([]string, len(b), len(b))
	for i, b := range b {
		names[i] = n.Name(b)
	}
	return strings.Join(names, sep)
}
