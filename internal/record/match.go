package record

import "strings"

// MatchScope reports whether recordScope matches the query pattern.
//
// The match rule is deliberately simple and is the v0 contract:
//   - exact equality, OR
//   - recordScope is a strict descendant of pattern, i.e. it starts
//     with pattern and the next character is a path separator.
//
// "src/foo" matches scope "src/foo" and "src/foo/bar.go". It does
// not match "src/foobar". A trailing slash on either side is
// tolerated but ignored. An empty pattern matches nothing (the
// caller is expected to validate the pattern before calling).
//
// Glob-style patterns in record scopes ("src/**") are not
// interpreted here. The scope field is treated as an opaque string
// on both sides; the only special character is the path
// separator.
func MatchScope(recordScope, pattern string) bool {
	recordScope = strings.TrimRight(recordScope, "/")
	pattern = strings.TrimRight(pattern, "/")
	if pattern == "" {
		return false
	}
	if recordScope == pattern {
		return true
	}
	return strings.HasPrefix(recordScope, pattern+"/")
}
