// Package storage is the test surface for the storage package.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/SincereMa/cortex-sidemark/internal/record"
)

// newStore returns a Store rooted at a per-test temp directory.
// t.TempDir handles cleanup.
func newStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	return NewStore(dir)
}

// sampleRecord returns a record of kind k with the kind-specific
// timestamp fields populated so it passes the schema in a
// hypothetical round-trip. It is intentionally minimal; the
// schema tests in package schema cover field-by-field cases. The
// id is derived from the test name and a counter so two records
// of the same kind do not collide on disk.
func sampleRecord(t *testing.T, k record.Kind) *record.Record {
	t.Helper()
	sampleCounter++
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r := &record.Record{
		ID:             fmt.Sprintf("01HW4F2N8X2JZPM7Q3S5V0K1A%d", sampleCounter),
		Kind:           k,
		Scope:          "src/foo",
		Subject:        "Hello World",
		Reason:         "test",
		SourceType:     record.SourceHuman,
		Author:         "tester",
		CreatedAt:      now,
		LastVerifiedAt: now,
		Status:         "active",
	}
	if k == record.KindDecision {
		d := now
		r.DecidedAt = &d
	}
	if k == record.KindExperiment {
		s := now
		r.StartedAt = &s
	}
	if k == record.KindIncident {
		o := now
		r.OccurredAt = &o
	}
	return r
}

// sampleCounter is incremented by sampleRecord so successive
// records of the same kind get distinct IDs.
var sampleCounter int

// TestWriteAndRead confirms a written record can be read back
// with its fields preserved.
func TestWriteAndRead(t *testing.T) {
	s := newStore(t)
	r := sampleRecord(t, record.KindDecision)

	path, err := s.Write(r)
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(path) {
		t.Errorf("expected absolute path, got %q", path)
	}

	got, err := s.Read(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != r.ID {
		t.Errorf("id mismatch: %q vs %q", got.ID, r.ID)
	}
	if got.Subject != r.Subject {
		t.Errorf("subject mismatch: %q vs %q", got.Subject, r.Subject)
	}
}

// TestWriteIsAtomic confirms the temporary write file does not
// survive a successful Write.
func TestWriteIsAtomic(t *testing.T) {
	s := newStore(t)
	r := sampleRecord(t, record.KindConstraint)
	path, err := s.Write(r)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("expected tmp to be cleaned up, got err=%v", err)
	}
}

// TestListByKind confirms List returns only files of the
// requested kind and tolerates an empty kind directory.
func TestListByKind(t *testing.T) {
	s := newStore(t)
	mustWrite := func(k record.Kind) {
		t.Helper()
		if _, err := s.Write(sampleRecord(t, k)); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite(record.KindDecision)
	mustWrite(record.KindDecision)
	mustWrite(record.KindConstraint)

	decisions, err := s.List(record.KindDecision)
	if err != nil {
		t.Fatal(err)
	}
	if len(decisions) != 2 {
		t.Errorf("expected 2 decisions, got %d", len(decisions))
	}

	constraints, err := s.List(record.KindConstraint)
	if err != nil {
		t.Fatal(err)
	}
	if len(constraints) != 1 {
		t.Errorf("expected 1 constraint, got %d", len(constraints))
	}

	if _, err := s.List(record.KindSignal); err != nil {
		t.Errorf("empty kind should not error: %v", err)
	}
}

// TestWriteRejectsInvalidKind confirms Write fails fast on a
// record whose Kind is not in the recognized set.
func TestWriteRejectsInvalidKind(t *testing.T) {
	s := newStore(t)
	r := sampleRecord(t, record.KindDecision)
	r.Kind = record.Kind("garbage")
	if _, err := s.Write(r); err == nil {
		t.Fatal("expected error for invalid kind")
	}
}

// TestRoundTripJSON confirms that pointer-typed optional fields
// (severity, valid_until) survive a JSON round-trip with their
// values intact.
func TestRoundTripJSON(t *testing.T) {
	s := newStore(t)
	r := sampleRecord(t, record.KindDecision)
	r.Tags = []string{"foo", "bar"}
	r.Evidence = []string{"https://example.com"}
	sev := record.SeverityHard
	r.Severity = &sev
	v := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)
	r.ValidUntil = &v

	path, err := s.Write(r)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var got record.Record
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Severity == nil || *got.Severity != sev {
		t.Errorf("severity not preserved: %+v", got.Severity)
	}
	if got.ValidUntil == nil || !got.ValidUntil.Equal(v) {
		t.Errorf("valid_until not preserved: %+v", got.ValidUntil)
	}
}

// TestListAllOrdering confirms ListAll returns records across all
// kinds sorted by CreatedAt descending.
func TestListAllOrdering(t *testing.T) {
	s := newStore(t)
	earlier := sampleRecord(t, record.KindDecision)
	earlier.CreatedAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	earlier.LastVerifiedAt = earlier.CreatedAt
	mustWrite(t, s, earlier)

	later := sampleRecord(t, record.KindConstraint)
	later.CreatedAt = time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	later.LastVerifiedAt = later.CreatedAt
	mustWrite(t, s, later)

	got, err := s.ListAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 records, got %d", len(got))
	}
	if !got[0].CreatedAt.After(got[1].CreatedAt) {
		t.Errorf("expected newest first; got %v before %v", got[0].CreatedAt, got[1].CreatedAt)
	}
}

// TestListKindFiltering confirms ListKind returns only records of
// the requested kind.
func TestListKindFiltering(t *testing.T) {
	s := newStore(t)
	mustWrite(t, s, sampleRecord(t, record.KindDecision))
	mustWrite(t, s, sampleRecord(t, record.KindConstraint))
	mustWrite(t, s, sampleRecord(t, record.KindDecision))

	decisions, err := s.ListKind(record.KindDecision)
	if err != nil {
		t.Fatal(err)
	}
	if len(decisions) != 2 {
		t.Errorf("expected 2 decisions, got %d", len(decisions))
	}
	for _, r := range decisions {
		if r.Kind != record.KindDecision {
			t.Errorf("kind filter leaked: got %q", r.Kind)
		}
	}
}

// TestGetExact confirms Get returns the matching record by exact
// id.
func TestGetExact(t *testing.T) {
	s := newStore(t)
	want := sampleRecord(t, record.KindDecision)
	mustWrite(t, s, want)

	got, err := s.Get(want.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != want.ID {
		t.Errorf("id mismatch: %q vs %q", got.ID, want.ID)
	}
}

// TestGetPrefix confirms Get falls back to prefix match when no
// exact id exists.
func TestGetPrefix(t *testing.T) {
	s := newStore(t)
	r := sampleRecord(t, record.KindConstraint)
	mustWrite(t, s, r)

	// Use the first 8 chars of the ULID as a prefix.
	got, err := s.Get(r.ID[:8])
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != r.ID {
		t.Errorf("id mismatch: %q vs %q", got.ID, r.ID)
	}
}

// TestGetPrefixAmbiguous confirms Get reports an error when the
// prefix matches more than one record.
func TestGetPrefixAmbiguous(t *testing.T) {
	s := newStore(t)
	a := sampleRecord(t, record.KindDecision)
	b := sampleRecord(t, record.KindConstraint)
	// Force the two ids to share a prefix by writing the same id
	// into two different kind directories; the kind dir is part
	// of the path, but findPrefix only inspects the basename, so
	// the file names will both start with the same prefix.
	a.ID = "01HWAMBIGUOUS00000000000000"
	b.ID = "01HWAMBIGUOUS00000000000001"
	mustWrite(t, s, a)
	mustWrite(t, s, b)

	_, err := s.Get("01HWAMBIGUOUS")
	if err == nil {
		t.Fatal("expected ambiguity error, got nil")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("error should mention ambiguity, got: %v", err)
	}
}

// TestGetMissing confirms Get reports a clean "no record" error
// when nothing matches.
func TestGetMissing(t *testing.T) {
	s := newStore(t)
	_, err := s.Get("01ZZZZZZZZZZZZZZZZZZZZZZZZ")
	if err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
}

// mustWrite is a test helper that writes r and fails the test on
// any error.
func mustWrite(t *testing.T, s *Store, r *record.Record) {
	t.Helper()
	if _, err := s.Write(r); err != nil {
		t.Fatal(err)
	}
}
