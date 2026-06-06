// Package cortex is the test surface for the get subcommand.
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

// seedRecord writes a single decision record to the store under
// cortexDir and returns the id used. The id is the canonical
// ULID-shaped string the other tests reuse.
func seedRecord(t *testing.T, cortexDir string) string {
	t.Helper()
	now := "2026-01-01T00:00:00Z"
	r := &record.Record{
		ID:             "01HW4F2N8X2JZPM7Q3S5V0K1A1",
		Kind:           record.KindDecision,
		Scope:          "src/foo",
		Subject:        "Use ULID for record ids",
		Reason:         "Lexicographic time-sortable ids make logs and listings easier.",
		SourceType:     record.SourceHuman,
		Author:         "tester",
		CreatedAt:      mustTime(t, now),
		LastVerifiedAt: mustTime(t, now),
		Status:         "active",
		DecidedAt:      timePtr(t, now),
	}
	s := storage.NewStore(cortexDir)
	if _, err := s.Write(r); err != nil {
		t.Fatal(err)
	}
	return r.ID
}

// mustTime parses a fixed RFC3339 string into a time.Time and
// fails the test on error.
func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	v, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse time %q: %v", s, err)
	}
	return v
}

// timePtr returns a pointer to a parsed time, failing the test
// on parse error.
func timePtr(t *testing.T, s string) *time.Time {
	t.Helper()
	v := mustTime(t, s)
	return &v
}

// TestGetJSON confirms `cortex get <id>` prints the record as
// indented JSON.
func TestGetJSON(t *testing.T) {
	root := t.TempDir()
	cortexDir := filepath.Join(root, ".cortex")
	if err := os.MkdirAll(cortexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	id := seedRecord(t, cortexDir)

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"get", "--root", cortexDir, id})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("get returned error: %v\n%s", err, out.String())
	}
	var got record.Record
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid record JSON: %v\n%s", err, out.String())
	}
	if got.ID != id {
		t.Errorf("id mismatch: %q vs %q", got.ID, id)
	}
}

// TestGetHuman confirms --human prints a one-line summary.
func TestGetHuman(t *testing.T) {
	root := t.TempDir()
	cortexDir := filepath.Join(root, ".cortex")
	if err := os.MkdirAll(cortexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	id := seedRecord(t, cortexDir)

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"get", "--human", "--root", cortexDir, id})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("get returned error: %v", err)
	}
	if !strings.Contains(out.String(), id) {
		t.Errorf("expected id in summary, got: %q", out.String())
	}
	if !strings.Contains(out.String(), "decision") {
		t.Errorf("expected kind in summary, got: %q", out.String())
	}
}

// TestGetPrefix confirms a prefix of the id resolves to the
// record.
func TestGetPrefix(t *testing.T) {
	root := t.TempDir()
	cortexDir := filepath.Join(root, ".cortex")
	if err := os.MkdirAll(cortexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	id := seedRecord(t, cortexDir)

	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"get", "--root", cortexDir, id[:8]})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("get returned error: %v", err)
	}
	if !strings.Contains(out.String(), id) {
		t.Errorf("expected full id in output, got: %q", out.String())
	}
}

// TestGetMissing confirms a non-existent id produces an error.
func TestGetMissing(t *testing.T) {
	root := t.TempDir()
	cortexDir := filepath.Join(root, ".cortex")
	if err := os.MkdirAll(cortexDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"get", "--root", cortexDir, "01ZZZZZZZZZZZZZZZZZZZZZZZZ"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing id, got nil")
	}
}
