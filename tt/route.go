package tt

import (
	"regexp"
	"strings"

	"github.com/gregoryv/mq"
)

func NewRoute(filter string, handlers ...mq.Handler) *Route {
	r := Route{filter: filter, handlers: handlers}
	if i := strings.Index(filter, "/#"); i > 0 {
		r.prefix = filter[:i]
	}
	if strings.Contains(filter, "+") {
		f := filter
		if r.prefix != "" {
			f = r.prefix
		}
		f = strings.ReplaceAll(f, "+", `(\w+)`)
		r.rx = regexp.MustCompile(f)
	}
	return &r
}

type Route struct {
	filter   string
	prefix   string
	rx       *regexp.Regexp
	handlers []mq.Handler
}

func (r *Route) String() string {
	return r.filter
}

func (r *Route) Match(name string) ([]string, bool) {
	switch {
	case r.filter == "#":
		return nil, true

	case r.prefix != "": // filter contains a #

		if r.rx != nil { // and prefix contains a +
			words := r.rx.FindAllStringSubmatch(name, -1)
			return words[0][1:], len(words) > 0
		}
		return nil, strings.HasPrefix(name, r.prefix)

	case r.rx != nil:
		words := r.rx.FindAllStringSubmatch(name, -1)
		return words[0][1:], len(words) > 0

	case name == r.filter:
		return nil, true
	}

	return nil, false
}
