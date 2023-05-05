// Package pathmatch allows you to quickly parse and match URL paths.
//
// Path expression are a composition of different segments.
//
// Parameterized segments start with a colon followed by a name
//
//	path			string		result
//	/foo/:name 		/foo 		nil
//	/foo/:name 		/foo/bar  	{"name": "bar"}
//
// Mixed segments can contain static and variable parts
// The end of a parameter name is detected by the following special characters: .?=&#:
// A suffix must be set, in order to match parameters, which are followed by an alphanumeric sequence,
//
//	path					string				result									options
//	/index.:ext?:p1=:v1		/index.html?x=1		{"ext": "html", "p1": "x", "v1": "1"}	none
//	/{start}def				/abcdef				{"start": "abc"}						prefix: "{", suffix: "}"
//
// The wildcard * matches one or more segments.
//
//	path	string		result
//	/* 		/foo 		{"$1": "foo"}
//	/* 		/foo/bar  	{"$1": "foo/bar"}
package pathmatch

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Match map[string]string

type savePoint struct {
	i           int
	sIndex      int
	searchStart int
	valid       bool
}

type Path struct {
	path       string
	Seperator  string
	Prefix     string
	Suffix     string
	Wildcard   string
	Segments   []ISegment
	match      Match
	save       *savePoint
	equalCheck bool
}

var except = regexp.MustCompile(`[^.?=&#:]+`)

// Compile parses a path expression and returns a Path if successful
func Compile(path string, options ...Option) (*Path, error) {
	p := &Path{path, "/", ":", "", "*", []ISegment{}, make(Match, 0), &savePoint{}, false}

	for _, option := range options {
		if err := option(p); err != nil {
			return nil, err
		}
	}

	unnamed := 0
	strSegments := strings.Split(path, p.Seperator)
	for _, strSeg := range strSegments {
		if strSeg == p.Wildcard {
			key := "$" + strconv.Itoa(unnamed)
			unnamed++
			p.Segments = append(p.Segments, newWildcardSegment(key))
		} else if iPrefix := strings.Index(strSeg, p.Prefix); iPrefix != -1 {

			var key string
			keyLocs := []int{} //locations including suffix and prefix
			keys := []string{}

			for iPrefix != -1 {
				keyStart := iPrefix + len(p.Prefix)
				if p.Suffix != "" {
					iSuffix := strings.Index(strSeg[iPrefix:], p.Suffix)
					if iSuffix == -1 {
						return nil, fmt.Errorf("pathmatch: %s, suffix \"%s\" not found", strSeg, p.Suffix)
					}
					key = strSeg[keyStart : iPrefix+iSuffix]
					keyLocs = append(keyLocs, iPrefix, iPrefix+iSuffix+len(p.Suffix))
				} else {
					keyLoc := except.FindStringIndex(strSeg[keyStart:])
					if keyLoc == nil || keyLoc[0] != 0 {
						return nil, fmt.Errorf("pathmatch: %s, prefix \"%s\" must be followed by name", strSeg, p.Suffix)
					}
					key = strSeg[keyStart : keyStart+keyLoc[1]]
					keyLocs = append(keyLocs, iPrefix, keyStart+keyLoc[1])
				}
				keys = append(keys, key)
				iPrefix = strings.Index(strSeg[keyStart:], p.Prefix)
				if iPrefix == -1 {
					break
				}
				iPrefix += keyStart
			}

			if len(keyLocs) == 2 && keyLocs[1]-keyLocs[0] == len(strSeg) {
				p.Segments = append(p.Segments, newParamSegment(key, p.equalCheck))
				continue
			}

			mixed, err := newMixedSegment(strSeg, keys, keyLocs)
			if err != nil {
				return nil, err
			}
			p.Segments = append(p.Segments, mixed)
		} else {
			p.Segments = append(p.Segments, newStaticSegment(strSeg))
		}
	}

	return p, nil
}

// Match returns true if s and p match
func (p *Path) Match(s string) bool {
	m := p.getMatch(s, false || p.equalCheck)
	return m != nil
}

// FindSubmatch returns a map with the values of parameterized segments, if s and p match
// Otherwise nil is returned
// Wildcard segments are named $0, $1, ...
func (p *Path) FindSubmatch(s string) Match {
	return p.getMatch(s, true)
}

func sliceSegment(s string, sep string, start int, offset int) (string, bool) {
	str := s[start:]
	i := strings.Index(str[offset:], sep)
	if i == -1 {
		return str, true
	}
	return str[:i+offset], false
}

func segmentLen(s string, sep string, done bool) int {
	if done {
		return len(s)
	}
	return len(s) + len(sep)
}

func (p *Path) getMatch(s string, capture bool) Match {
	draft := newMatchDraft(capture, p.match)

	sIndex := 0
	searchStart := 0

	for i := 0; draft != nil && i < len(p.Segments); i++ {
		seg := p.Segments[i]

		str, done := sliceSegment(s, p.Seperator, sIndex, searchStart)
		if done && len(p.Segments)-1 != i {
			return nil
		}

		if seg.Multiple() {

			if len(p.Segments)-1 == i {
				draft = seg.Match(draft, s[sIndex:])
				sIndex = len(s)
				break
			}

			if p.save.valid && p.save.i == i {
				p.save.searchStart = segmentLen(str, p.Seperator, done)
			} else {
				p.save.i = i
				p.save.sIndex = sIndex
				p.save.searchStart = segmentLen(str, p.Seperator, done)
				p.save.valid = true
			}
		}

		m := seg.Match(draft, str)
		if m == nil && p.save.valid {
			i = p.save.i - 1
			sIndex = p.save.sIndex
			searchStart = p.save.searchStart
			continue
		}

		draft = m
		sIndex += segmentLen(str, p.Seperator, done)
		searchStart = 0

		if len(p.Segments)-1 == i && !done {
			return nil
		}
	}
	if draft == nil || len(s) != sIndex {
		return nil
	}
	return draft.match
}

// IsStatic returns true if p only contains static segments
func (p *Path) IsStatic() bool {
	for _, seg := range p.Segments {
		if seg.Type() != Static {
			return false
		}
	}
	return true
}
