package tt

import (
	"testing"
)

func TestRoute(t *testing.T) {
	// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901241
	spec := []string{
		"sport/tennis/player1",
		"sport/tennis/player1/ranking",
		"sport/tennis/player1/score/wimbledon",
	}

	cases := []struct {
		expMatch bool
		names    []string
		*Route
		expWords []string
	}{
		{true, spec, NewRoute("sport/tennis/player1/#"), nil},
		{true, spec, NewRoute("sport/#"), nil},
		{true, spec, NewRoute("#"), nil},
		{true, spec, NewRoute("+/tennis/#"), []string{"sport"}},

		{true, []string{"a/b/c"}, NewRoute("a/+/+"), []string{"b", "c"}},
		{true, []string{"a/b/c"}, NewRoute("a/+/c"), []string{"b"}},

		{false, spec, NewRoute("+"), nil},
		{false, spec, NewRoute("tennis/player1/#"), nil},
		{false, spec, NewRoute("sport/tennis#"), nil},
	}

	for _, c := range cases {
		for _, name := range c.names {
			words, match := c.Route.Match(name)

			if !equal(words, c.expWords) || match != c.expMatch {
				t.Errorf("%s %s exp:%v got:%v %q",
					name, c.Route, c.expMatch, match, words,
				)
			}

			if v := c.Route.Filter(); v == "" {
				t.Error("no subscription")
			}
		}
	}

	// check String
	if v := NewRoute("sport/#").String(); v != "sport/#" {
		t.Error("Route.String missing filter", v)
	}
}

func equal[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
