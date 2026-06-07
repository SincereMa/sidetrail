// Package sidetrail is the test surface for the ask subcommand.
package sidetrail

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/SincereMa/sidetrail/internal/record"
	"github.com/SincereMa/sidetrail/internal/storage"
)

// seedAskFixture writes a small mixed-scope set of records to
// the store and returns the store directory.
func seedAskFixture(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), storeDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	s := storage.NewStore(dir)

	writeAt := func(id, scope, kind, tag string, when time.Time) {
		t.Helper()
		r := &record.Record{
			ID:             id,
			Kind:           record.Kind(kind),
			Scope:          scope,
			Subject:        "fixture",
			Reason:         "fixture",
			SourceType:     record.SourceHuman,
			Author:         "tester",
			CreatedAt:      when,
			LastVerifiedAt: when,
			Status:         "active",
		}
		if tag != "" {
			r.Tags = []string{tag}
		}
		if _, err := s.Write(r); err != nil {
			t.Fatal(err)
		}
	}

	writeAt("01HWAAAAAAAAAAAA0000000001", "src/foo", "decision", "auth", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	writeAt("01HWAAAAAAAAAAAA0000000002", "src/foo/bar.go", "constraint", "auth", time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))
	writeAt("01HWAAAAAAAAAAAA0000000003", "src/foobar", "signal", "", time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))
	writeAt("01HWAAAAAAAAAAAA0000000004", "auth", "decision", "billing", time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC))
	return dir
}

// TestAskScopePattern confirms `sidetrail ask --scope src/foo`
// returns records whose scope matches by exact or strict
// descendant.
func TestAskScopePattern(t *testing.T) {
	dir := seedAskFixture(t)
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"ask", "--root", dir, "--scope", "src/foo", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("ask returned error: %v\n%s", err, out.String())
	}
	var recs []*record.Record
	if err := json.Unmarshal(out.Bytes(), &recs); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out.String())
	}
	if len(recs) != 2 {
		t.Fatalf("expected 2 records for src/foo, got %d: %s", len(recs), out.String())
	}
	gotScopes := map[string]bool{}
	for _, r := range recs {
		gotScopes[r.Scope] = true
	}
	if !gotScopes["src/foo"] {
		t.Errorf("expected src/foo in scopes, got %v", gotScopes)
	}
	if !gotScopes["src/foo/bar.go"] {
		t.Errorf("expected src/foo/bar.go (descendant) in scopes, got %v", gotScopes)
	}
	if gotScopes["src/foobar"] {
		t.Errorf("expected src/foobar to be excluded, got %v", gotScopes)
	}
	if gotScopes["auth"] {
		t.Errorf("expected auth to be excluded, got %v", gotScopes)
	}
}

// TestAskKindTagFilter confirms --kind and --tag compose with
// --scope.
func TestAskKindTagFilter(t *testing.T) {
	dir := seedAskFixture(t)
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"ask", "--root", dir, "--scope", "src/foo", "--kind", "decision", "--tag", "auth", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("ask returned error: %v\n%s", err, out.String())
	}
	var recs []*record.Record
	if err := json.Unmarshal(out.Bytes(), &recs); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out.String())
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 record (decision + auth under src/foo), got %d: %s", len(recs), out.String())
	}
	if recs[0].Kind != record.KindDecision {
		t.Errorf("expected decision kind, got %q", recs[0].Kind)
	}
}

// TestAskJSON confirms --json emits a parseable array.
func TestAskJSON(t *testing.T) {
	dir := seedAskFixture(t)
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"ask", "--root", dir, "--scope", "src/foo", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("ask returned error: %v", err)
	}
	var recs []*record.Record
	if err := json.Unmarshal(out.Bytes(), &recs); err != nil {
		t.Fatalf("output is not a JSON array of records: %v\n%s", err, out.String())
	}
	if len(recs) != 2 {
		t.Errorf("expected 2 records for src/foo, got %d", len(recs))
	}
}

// TestAskMissingScope confirms --scope is required.
func TestAskMissingScope(t *testing.T) {
	dir := seedAskFixture(t)
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"ask", "--root", dir})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --scope, got nil")
	}
	if !strings.Contains(err.Error(), "scope") {
		t.Errorf("error should mention scope, got: %v", err)
	}
}
