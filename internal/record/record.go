// Package record defines the canonical in-memory and on-disk shape
// of a SideTrail record, plus helpers for ID and slug
// generation.
package record

import (
	"fmt"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

// Kind enumerates the five product-surface categories defined in
// ADR-0004. New kinds enter by ADR.
type Kind string

// The five record kinds.
const (
	KindDecision   Kind = "decision"
	KindConstraint Kind = "constraint"
	KindSignal     Kind = "signal"
	KindExperiment Kind = "experiment"
	KindIncident   Kind = "incident"
)

// Valid reports whether k is one of the recognized Kind values.
func (k Kind) Valid() bool {
	switch k {
	case KindDecision, KindConstraint, KindSignal, KindExperiment, KindIncident:
		return true
	}
	return false
}

// SourceType records who (or what) produced a record. The read
// layer down-weights non-human sources per ADR-0001.
type SourceType string

// The four source types.
const (
	SourceHuman          SourceType = "human"
	SourceAgentSuggested SourceType = "agent-suggested"
	SourceScrape         SourceType = "scrape"
	SourceDerived        SourceType = "derived"
)

// Valid reports whether s is one of the recognized SourceType values.
func (s SourceType) Valid() bool {
	switch s {
	case SourceHuman, SourceAgentSuggested, SourceScrape, SourceDerived:
		return true
	}
	return false
}

// Severity distinguishes hard constraints (real consequences on
// violation) from soft ones (team preferences or historical
// lessons). Per ADR-0002.
type Severity string

// The two severity levels.
const (
	SeverityHard Severity = "hard"
	SeveritySoft Severity = "soft"
)

// Valid reports whether s is one of the recognized Severity values.
func (s Severity) Valid() bool {
	return s == SeverityHard || s == SeveritySoft
}

// Record is the canonical in-memory shape of one entry in
// .sidetrail/. It is the union of all kind-specific fields defined
// in ADR-0001, ADR-0002, and ADR-0004; only the fields relevant
// to a record's kind are populated.
//
// Optional fields use pointer or empty-slice types so the
// distinction between "absent" and "zero value" survives JSON
// round-trips.
type Record struct {
	ID                   string     `json:"id"`
	Kind                 Kind       `json:"kind"`
	Scope                string     `json:"scope"`
	Subject              string     `json:"subject"`
	Body                 string     `json:"body,omitempty"`
	Reason               string     `json:"reason"`
	Evidence             []string   `json:"evidence,omitempty"`
	SourceType           SourceType `json:"source_type"`
	Author               string     `json:"author"`
	CreatedAt            time.Time  `json:"created_at"`
	LastVerifiedAt       time.Time  `json:"last_verified_at"`
	Status               string     `json:"status"`
	Supersedes           string     `json:"supersedes,omitempty"`
	SupersededBy         string     `json:"superseded_by,omitempty"`
	Tags                 []string   `json:"tags,omitempty"`
	Severity             *Severity  `json:"severity,omitempty"`
	ValidUntil           *time.Time `json:"valid_until,omitempty"`
	DecidedAt            *time.Time `json:"decided_at,omitempty"`
	StartedAt            *time.Time `json:"started_at,omitempty"`
	EndedAt              *time.Time `json:"ended_at,omitempty"`
	OccurredAt           *time.Time `json:"occurred_at,omitempty"`
	ResolvedAt           *time.Time `json:"resolved_at,omitempty"`
	RelatedTo            []string   `json:"related_to,omitempty"`
	RejectedAlternatives []string   `json:"rejected_alternatives,omitempty"`
}

// NewID returns a fresh ULID-encoded record identifier. The
// returned value is lexicographically sortable by creation time.
func NewID() (string, error) {
	id, err := ulid.New(ulid.Timestamp(time.Now()), ulid.Monotonic(randReader{}, 0))
	if err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return id.String(), nil
}

// Slug returns a filesystem-safe slug derived from s. The slug is
// lowercase, dash-separated, and capped at 48 characters so it
// composes with a record ID into a filename that stays readable.
func Slug(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	out := strings.TrimRight(b.String(), "-")
	if out == "" {
		return "record"
	}
	if len(out) > 48 {
		out = out[:48]
		out = strings.TrimRight(out, "-")
	}
	return out
}
