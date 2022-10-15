package tt

import (
	"testing"
)

func TestRoute(t *testing.T) {
	// https://docs.oasis-open.org/mqtt/mqtt/v5.0/os/mqtt-v5.0-os.html#_Toc3901241
	names := []string{
		"sport/tennis/player1",
		"sport/tennis/player1/ranking",
		"sport/tennis/player1/score/wimbledon",
	}

	cases := []struct {
		expMatch bool
		*Route
		expWords []string
	}{
		{true, NewRoute("sport/tennis/player1/#"), nil},
		{true, NewRoute("sport/#"), nil},
		{true, NewRoute("#"), nil},
		{true, NewRoute("+"), []string{"sport"}},
		{true, NewRoute("+/tennis/#"), []string{"sport"}},
		{false, NewRoute("tennis/player1/#"), nil},
		{false, NewRoute("sport/tennis#"), nil},
	}

	for _, c := range cases {
		for _, name := range names {
			words, match := c.Route.Match(name)
			if match != c.expMatch {
				t.Error(c.expMatch, c.Route, "got", match)
			}

			if !equal(words, c.expWords) {
				t.Log(name)
				t.Error(c.Route, words)
			}
		}
	}

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
