package internal

import "strings"

type Dict map[byte]string

func (n Dict) Join(sep string, b []byte) string {
	if len(b) == 0 {
		return ""
	}
	names := make([]string, len(b), len(b))
	for i, b := range b {
		names[i] = n[b]
	}
	return strings.Join(names, sep)
}
