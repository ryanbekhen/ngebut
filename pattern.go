package ngebut

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

// A pattern is something that can be matched against an HTTP request.
type pattern struct {
	str      string // original string
	method   string
	host     string
	segments []segment
}

// A segment is a pattern piece that matches one or more path segments, or a trailing slash.
type segment struct {
	s     string // literal or wildcard name or "/" for "/{$}".
	wild  bool
	multi bool // "..." wildcard
}

func parsePattern(s string) (_ *pattern, err error) {
	if s == "" {
		return nil, errors.New("empty pattern")
	}
	off := 0 // offset into string
	defer func() {
		if err != nil {
			err = fmt.Errorf("at offset %d: %w", off, err)
		}
	}()

	method, rest, found := strings.Cut(s, " ")
	if !found {
		rest = method
		method = ""
	}
	if method != "" && !validMethod(method) {
		return nil, fmt.Errorf("invalid method %q", method)
	}
	p := &pattern{str: s, method: method}

	if found {
		off = len(method) + 1
	}
	i := strings.IndexByte(rest, '/')
	if i < 0 {
		return nil, errors.New("host/path missing /")
	}
	p.host = rest[:i]
	rest = rest[i:]
	if j := strings.IndexByte(p.host, '{'); j >= 0 {
		off += j
		return nil, errors.New("host contains '{' (missing initial '/'?)")
	}
	// At this point, rest is the path.
	off += i

	seenNames := map[string]bool{} // remember wildcard names to catch dups
	for len(rest) > 0 {
		rest = rest[1:]
		off = len(s) - len(rest)
		if len(rest) == 0 {
			p.segments = append(p.segments, segment{wild: true, multi: true})
			break
		}
		i := strings.IndexByte(rest, '/')
		if i < 0 {
			i = len(rest)
		}
		var seg string
		seg, rest = rest[:i], rest[i:]
		if i := strings.IndexByte(seg, '{'); i < 0 {
			seg = pathUnescape(seg)
			p.segments = append(p.segments, segment{s: seg})
		} else {
			if i != 0 {
				return nil, errors.New("bad wildcard segment (must start with '{')")
			}
			if seg[len(seg)-1] != '}' {
				return nil, errors.New("bad wildcard segment (must end with '}')")
			}
			name := seg[1 : len(seg)-1]
			if name == "$" {
				if len(rest) != 0 {
					return nil, errors.New("{$} not at end")
				}
				p.segments = append(p.segments, segment{s: "/"})
				break
			}
			name, multi := strings.CutSuffix(name, "...")
			if multi && len(rest) != 0 {
				return nil, errors.New("{...} wildcard not at end")
			}
			if name == "" {
				return nil, errors.New("empty wildcard")
			}
			if !isValidWildcardName(name) {
				return nil, fmt.Errorf("bad wildcard name %q", name)
			}
			if seenNames[name] {
				return nil, fmt.Errorf("duplicate wildcard name %q", name)
			}
			seenNames[name] = true
			p.segments = append(p.segments, segment{s: name, wild: true, multi: multi})
		}
	}
	return p, nil
}

func validMethod(method string) bool {
	return method == "GET" || method == "POST" || method == "PUT" || method == "DELETE" || method == "HEAD" || method == "OPTIONS" || method == "PATCH"
}

func isValidWildcardName(s string) bool {
	if s == "" {
		return false
	}
	for i, c := range s {
		if !unicode.IsLetter(c) && c != '_' && (i == 0 || !unicode.IsDigit(c)) {
			return false
		}
	}
	return true
}

func pathUnescape(path string) string {
	u, err := url.PathUnescape(path)
	if err != nil {
		return path
	}
	return u
}

type relationship string

const (
	equivalent   relationship = "equivalent"
	moreGeneral  relationship = "moreGeneral"
	moreSpecific relationship = "moreSpecific"
	disjoint     relationship = "disjoint"
	overlaps     relationship = "overlaps"
)

func (p1 *pattern) comparePathsAndMethods(p2 *pattern) relationship {
	mrel := p1.compareMethods(p2)
	if mrel == disjoint {
		return disjoint
	}
	prel := p1.comparePaths(p2)
	return combineRelationships(mrel, prel)
}

func (p1 *pattern) compareMethods(p2 *pattern) relationship {
	if p1.method == p2.method {
		return equivalent
	}
	if p1.method == "" {
		return moreGeneral
	}
	if p2.method == "" {
		return moreSpecific
	}
	if p1.method == "GET" && p2.method == "HEAD" {
		return moreGeneral
	}
	if p2.method == "GET" && p1.method == "HEAD" {
		return moreSpecific
	}
	return disjoint
}

func (p1 *pattern) comparePaths(p2 *pattern) relationship {
	if len(p1.segments) != len(p2.segments) && !p1.lastSegment().multi && !p2.lastSegment().multi {
		return disjoint
	}

	var segs1, segs2 []segment
	rel := equivalent
	for segs1, segs2 = p1.segments, p2.segments; len(segs1) > 0 && len(segs2) > 0; segs1, segs2 = segs1[1:], segs2[1:] {
		rel = combineRelationships(rel, compareSegments(segs1[0], segs2[0]))
		if rel == disjoint {
			return rel
		}
	}

	if len(segs1) == 0 && len(segs2) == 0 {
		return rel
	}
	if len(segs1) < len(segs2) && p1.lastSegment().multi {
		return combineRelationships(rel, moreGeneral)
	}
	if len(segs2) < len(segs1) && p2.lastSegment().multi {
		return combineRelationships(rel, moreSpecific)
	}
	return disjoint
}

func compareSegments(s1, s2 segment) relationship {
	if s1.multi && s2.multi {
		return equivalent
	}
	if s1.multi {
		return moreGeneral
	}
	if s2.multi {
		return moreSpecific
	}
	if s1.wild && s2.wild {
		return equivalent
	}
	if s1.wild {
		if s2.s == "/" {
			return disjoint
		}
		return moreGeneral
	}
	if s2.wild {
		if s1.s == "/" {
			return disjoint
		}
		return moreSpecific
	}
	if s1.s == s2.s {
		return equivalent
	}
	return disjoint
}

func combineRelationships(r1, r2 relationship) relationship {
	switch r1 {
	case equivalent:
		return r2
	case disjoint:
		return disjoint
	case overlaps:
		if r2 == disjoint {
			return disjoint
		}
		return overlaps
	case moreGeneral, moreSpecific:
		switch r2 {
		case equivalent:
			return r1
		case inverseRelationship(r1):
			return overlaps
		default:
			return r2
		}
	default:
		panic(fmt.Sprintf("unknown relationship %q", r1))
	}
}

func inverseRelationship(r relationship) relationship {
	switch r {
	case moreSpecific:
		return moreGeneral
	case moreGeneral:
		return moreSpecific
	default:
		return r
	}
}

func (p *pattern) lastSegment() segment {
	return p.segments[len(p.segments)-1]
}

func (p *pattern) conflictsWith(other *pattern) bool {
	if p.method != other.method {
		return false
	}
	if p.host != other.host {
		return false
	}
	rel := p.comparePaths(other)
	return rel == equivalent || rel == overlaps
}

func (p *pattern) match(req *Request) bool {
	if p.method != "" && p.method != req.Method {
		return false
	}
	if p.host != "" && !strings.HasPrefix(req.Host, p.host) {
		return false
	}
	return p.matchPath(req.URL.Path)
}

func (p *pattern) matchPath(path string) bool {
	return strings.HasPrefix(path, p.str) && (path == p.str || p.lastSegment().multi)
}
