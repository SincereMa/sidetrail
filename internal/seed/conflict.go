// Package seed provides functions for bootstrapping SideTrail
// records from existing project documents.
package seed

import (
	"strings"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/storage"
)

// Conflict represents a candidate record that conflicts with an
// existing record.
type Conflict struct {
	Candidate *record.Record `json:"candidate"`
	Existing  *record.Record `json:"existing"`
}

// DetectConflicts compares candidate records against the store and
// returns conflicts (same kind + overlapping scope + similar
// subject) and non-conflicting records.
func DetectConflicts(store *storage.Store, candidates []*record.Record) ([]Conflict, []*record.Record) {
	existing, err := store.ListAll()
	if err != nil {
		return nil, candidates
	}

	var conflicts []Conflict
	var nonConflicting []*record.Record

	for _, c := range candidates {
		conflict := findConflict(c, existing)
		if conflict != nil {
			conflicts = append(conflicts, *conflict)
		} else {
			nonConflicting = append(nonConflicting, c)
		}
	}

	return conflicts, nonConflicting
}

// findConflict checks if candidate conflicts with any existing
// record.
func findConflict(candidate *record.Record, existing []*record.Record) *Conflict {
	for _, e := range existing {
		if candidate.Kind != e.Kind {
			continue
		}
		if !scopeOverlaps(candidate.Scope, e.Scope) {
			continue
		}
		if subjectsMatch(candidate.Subject, e.Subject) {
			return &Conflict{Candidate: candidate, Existing: e}
		}
	}
	return nil
}

// scopeOverlaps reports whether two scopes overlap (exact match or
// one is ancestor of other).
func scopeOverlaps(a, b string) bool {
	a = strings.TrimRight(a, "/")
	b = strings.TrimRight(b, "/")
	if a == b {
		return true
	}
	return strings.HasPrefix(a, b+"/") || strings.HasPrefix(b, a+"/")
}

// subjectsMatch reports whether two subjects are similar (case-
// insensitive prefix match).
func subjectsMatch(a, b string) bool {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))
	if a == b {
		return true
	}
	return strings.HasPrefix(a, b) || strings.HasPrefix(b, a)
}
