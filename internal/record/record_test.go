// Package record is the test surface for the record package.
package record

import (
	"strings"
	"testing"
	"time"
)

// TestSlug exercises Slug's lowercasing, dash separation, and
// fallback behavior on a table of representative inputs.
func TestSlug(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"Hello World", "hello-world"},
		{"  spaces   collapse  ", "spaces-collapse"},
		{"already-kebab", "already-kebab"},
		{"CamelCaseTest", "camelcasetest"},
		{"punct.uation!here", "punct-uation-here"},
		{"", "record"},
		{"中文", "record"},
		{"a-b-c-d-e", "a-b-c-d-e"},
	}
	for _, c := range cases {
		got := Slug(c.in)
		if got != c.want {
			t.Errorf("Slug(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// TestSlugLength verifies Slug caps its output at 48 characters
// and never leaves a trailing dash.
func TestSlugLength(t *testing.T) {
	s := Slug("this is a very long subject that should be truncated at some sensible boundary")
	if len(s) > 48 {
		t.Errorf("slug too long: %d chars (%q)", len(s), s)
	}
	if strings.HasSuffix(s, "-") {
		t.Errorf("slug ends with dash: %q", s)
	}
}

// TestKindValid confirms Kind.Valid accepts all five defined kinds
// and rejects unknown values.
func TestKindValid(t *testing.T) {
	valid := []Kind{KindDecision, KindConstraint, KindSignal, KindExperiment, KindIncident}
	for _, k := range valid {
		if !k.Valid() {
			t.Errorf("Kind(%q).Valid() = false, want true", k)
		}
	}
	invalid := []Kind{"", "unknown", "Decision", "CONSTRAINT"}
	for _, k := range invalid {
		if k.Valid() {
			t.Errorf("Kind(%q).Valid() = true, want false", k)
		}
	}
}

// TestSourceTypeValid confirms SourceType.Valid accepts all four
// defined source types and rejects unknown values.
func TestSourceTypeValid(t *testing.T) {
	valid := []SourceType{SourceHuman, SourceAgentSuggested, SourceScrape, SourceDerived}
	for _, s := range valid {
		if !s.Valid() {
			t.Errorf("SourceType(%q).Valid() = false, want true", s)
		}
	}
	invalid := []SourceType{"", "unknown", "Human", "AGENT"}
	for _, s := range invalid {
		if s.Valid() {
			t.Errorf("SourceType(%q).Valid() = true, want false", s)
		}
	}
}

// TestSeverityValid confirms Severity.Valid accepts hard and soft
// and rejects unknown values.
func TestSeverityValid(t *testing.T) {
	valid := []Severity{SeverityHard, SeveritySoft}
	for _, s := range valid {
		if !s.Valid() {
			t.Errorf("Severity(%q).Valid() = false, want true", s)
		}
	}
	invalid := []Severity{"", "medium", "critical", "HARD"}
	for _, s := range invalid {
		if s.Valid() {
			t.Errorf("Severity(%q).Valid() = true, want false", s)
		}
	}
}

// TestNewIDMonotonic verifies that two NewID calls separated by
// a measurable interval return strictly increasing IDs.
func TestNewIDMonotonic(t *testing.T) {
	a, err := NewID()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(2 * time.Millisecond)
	b, err := NewID()
	if err != nil {
		t.Fatal(err)
	}
	if a >= b {
		t.Errorf("ids not increasing: %q >= %q", a, b)
	}
}
