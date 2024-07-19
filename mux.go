package ngebut

import (
	"fmt"
	"sort"
)

type ServeMux struct {
	handlers map[*pattern]Handler
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		handlers: make(map[*pattern]Handler),
	}
}

func (mux *ServeMux) Handle(pattern string, handler Handler) {
	pat, err := parsePattern(pattern)
	if err != nil {
		panic(err)
	}

	for p := range mux.handlers {
		if p.conflictsWith(pat) {
			panic(fmt.Sprintf("pattern %q conflicts with existing pattern %q", pattern, p.str))
		}
	}

	mux.handlers[pat] = handler
}

func (mux *ServeMux) HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
	mux.Handle(pattern, HandlerFunc(handler))
}

func (mux *ServeMux) ServeHTTP(w ResponseWriter, r *Request) {
	var sortedPatterns []*pattern
	for p := range mux.handlers {
		sortedPatterns = append(sortedPatterns, p)
	}

	sort.Slice(sortedPatterns, func(i, j int) bool {
		return len(sortedPatterns[i].segments) > len(sortedPatterns[j].segments)
	})

	for _, p := range sortedPatterns {
		if p.match(r) {
			mux.handlers[p].ServeHTTP(w, r)
			return
		}
	}

	notFoundHandler(w, r)
}

func notFoundHandler(w ResponseWriter, r *Request) {
	w.WriteHeader(404)
	w.Write([]byte("404 page not found"))
}
