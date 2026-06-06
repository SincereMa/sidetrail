// Package record is the test surface for MatchScope.
package record

import "testing"

// TestMatchScope exercises the exact-match and descendant rules.
func TestMatchScope(t *testing.T) {
	cases := []struct {
		scope, pattern string
		want           bool
	}{
		// Exact match.
		{"src/foo", "src/foo", true},
		{"auth", "auth", true},
		// Descendant match.
		{"src/foo/bar.go", "src/foo", true},
		{"src/foo/bar/baz", "src/foo", true},
		// Trailing slash on the pattern is tolerated.
		{"src/foo", "src/foo/", true},
		{"src/foo/bar.go", "src/foo/", true},
		// "src/foobar" must NOT match "src/foo".
		{"src/foobar", "src/foo", false},
		// Empty pattern matches nothing.
		{"src/foo", "", false},
		{"", "", false},
		// Different areas.
		{"auth", "src/foo", false},
	}
	for _, c := range cases {
		got := MatchScope(c.scope, c.pattern)
		if got != c.want {
			t.Errorf("MatchScope(%q, %q) = %v, want %v", c.scope, c.pattern, got, c.want)
		}
	}
}
