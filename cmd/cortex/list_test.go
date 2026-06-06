// Package cortex is the test surface for the list subcommand.
package cortex

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/SincereMa/cortex-sidemark/internal/record"
	"github.com/SincereMa/cortex-sidemark/internal/storage"
)

// seedListFixture writes a small mixed-kind set of records to
// the store and returns the store directory.
func seedListFixture(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), ".cortex")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	s := storage.NewStore(dir)

	writeAt := func(id string, k record.Kind, subject string, when time.Time) {
		t.Helper()
		r := &record.Record{
			ID:             id,
			Kind:           k,
			Scope:          "src/foo",
			Subject:        subject,
			Reason:         "fixture",
			SourceType:     record.SourceHuman,
			Author:         "tester",
			CreatedAt:      when,
			LastVerifiedAt: when,
			Status:         "active",
		}
		if k == record.KindDecision {
			d := when
			r.DecidedAt = &d
		}
		if _, err := s.Write(r); err != nil {
			t.Fatal(err)
		}
	}

	writeAt("01HWAAAAAAAAAAAA0000000001", record.KindDecision, "Decision one", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	writeAt("01HWAAAAAAAAAAAA0000000002", record.KindDecision, "Decision two", time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	writeAt("01HWAAAAAAAAAAAA0000000003", record.KindConstraint, "Constraint one", time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))
	return dir
}

// TestListAllTable confirms the default table output contains a
// header and the expected ids.
func TestListAllTable(t *testing.T) {
	dir := seedListFixture(t)

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list", "--root", dir})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("list returned error: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "ID\tKIND\tSUBJECT") {
		t.Errorf("expected header in output, got: %q", out.String())
	}
	if !strings.Contains(out.String(), "Decision two") {
		t.Errorf("expected 'Decision two' in output, got: %q", out.String())
	}
	// Newest first: Decision two (June) must appear before
	// Constraint one (March).
	iTwo := strings.Index(out.String(), "Decision two")
	iConstraint := strings.Index(out.String(), "Constraint one")
	if iTwo < 0 || iConstraint < 0 || iTwo > iConstraint {
		t.Errorf("expected newest-first ordering; got:\n%s", out.String())
	}
}

// TestListKindFilter confirms --kind returns only that kind.
func TestListKindFilter(t *testing.T) {
	dir := seedListFixture(t)

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list", "--root", dir, "--kind", "decision"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("list returned error: %v", err)
	}
	if strings.Contains(out.String(), "Constraint one") {
		t.Errorf("expected no constraints, got: %q", out.String())
	}
	if !strings.Contains(out.String(), "Decision one") {
		t.Errorf("expected Decision one in output, got: %q", out.String())
	}
}

// TestListLimit confirms --limit caps the number of returned
// records.
func TestListLimit(t *testing.T) {
	dir := seedListFixture(t)

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list", "--root", dir, "--limit", "1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("list returned error: %v", err)
	}
	// Count non-header data lines.
	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	if len(lines) != 2 { // 1 header + 1 record
		t.Errorf("expected 2 lines (header + 1), got %d:\n%s", len(lines), out.String())
	}
}

// TestListJSON confirms --json emits a parseable JSON array.
func TestListJSON(t *testing.T) {
	dir := seedListFixture(t)

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list", "--root", dir, "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("list returned error: %v", err)
	}
	var recs []*record.Record
	if err := json.Unmarshal(out.Bytes(), &recs); err != nil {
		t.Fatalf("output is not a JSON array of records: %v\n%s", err, out.String())
	}
	if len(recs) != 3 {
		t.Errorf("expected 3 records, got %d", len(recs))
	}
}

// TestListUnknownKind confirms an unknown --kind is reported
// cleanly.
func TestListUnknownKind(t *testing.T) {
	dir := seedListFixture(t)
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"list", "--root", dir, "--kind", "banana"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for unknown kind, got nil")
	}
}
