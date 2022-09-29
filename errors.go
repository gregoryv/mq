package mq

import "fmt"

func unmarshalErr(v interface{}, ref string, err interface{}) *Malformed {
	e := newMalformed(v, ref, err)
	e.method = "unmarshal"
	return e
}

func newMalformed(v interface{}, ref string, err interface{}) *Malformed {
	var reason string
	switch e := err.(type) {
	case *Malformed:
		reason = e.reason
	case string:
		reason = e
	}
	// remove * from type name
	t := fmt.Sprintf("%T", v)
	if t[0] == '*' {
		t = t[1:]
	}
	return &Malformed{
		t:      t,
		ref:    ref,
		reason: reason,
	}
}

type Malformed struct {
	method string // fill or unmarshal
	t      string // the control packet
	ref    string
	reason string
}

func (e *Malformed) Error() string {
	if e.ref == "" {
		return fmt.Sprintf("malformed %s %s: %s", e.t, e.method, e.reason)
	}
	return fmt.Sprintf("malformed %s %s: %s %s", e.t, e.method, e.ref, e.reason)
}
