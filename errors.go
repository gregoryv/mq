package mq

import (
	"fmt"
	"strings"
)

func unmarshalErr(v interface{}, ref string, err interface{}) *Malformed {
	e := newMalformed(v, ref, err)
	e.method = "unmarshal"
	return e
}

func newMalformed(v interface{}, ref string, reason interface{}) *Malformed {
	var r string
	switch e := reason.(type) {
	case *Malformed:
		r = e.reason
	case string:
		r = e
	}
	return &Malformed{
		t:      fmt.Sprintf("%T", v),
		ref:    ref,
		reason: r,
	}
}

type Malformed struct {
	t      string // the control packet
	method string // fill or unmarshal
	ref    string
	reason string
}

func (e *Malformed) SetPacket(p Packet) {
	e.t = fmt.Sprintf("%T", p)
}

func (e *Malformed) SetReason(v string) { e.reason = v }

func (e *Malformed) Error() string {
	var buf strings.Builder
	buf.WriteString("malformed")
	add := func(v string) {
		if v == "" {
			return
		}
		buf.WriteString(" ")
		buf.WriteString(v)
	}
	add(e.t)
	add(e.method)
	buf.WriteString(":")
	add(e.ref)
	add(e.reason)
	return buf.String()
}
