// Package sidetrail is the test surface for the context subcommand.
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

// seedContextFixture writes records at three depths of the file
// tree plus a couple of unrelated scopes.
func seedContextFixture(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), storeDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	s := storage.NewStore(dir)

	writeAt := func(id, scope string) {
		t.Helper()
		r := &record.Record{
			ID:             id,
			Kind:           record.KindConstraint,
			Scope:          scope,
			Subject:        "fixture",
			Reason:         "fixture",
			SourceType:     record.SourceHuman,
			Author:         "tester",
			CreatedAt:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			LastVerifiedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			Status:         "active",
		}
		if _, err := s.Write(r); err != nil {
			t.Fatal(err)
		}
	}

	writeAt("01HWAAAAAAAAAAAA0000000001", "src/foo/bar/baz.go")
	writeAt("01HWAAAAAAAAAAAA0000000002", "src/foo/bar")
	writeAt("01HWAAAAAAAAAAAA0000000003", "src/foo")
	writeAt("01HWAAAAAAAAAAAA0000000004", "src/other")
	writeAt("01HWAAAAAAAAAAAA0000000005", "auth")
	return dir
}

// TestContextForFile confirms `sidetrail context --file <path>`
// returns records whose scope is the file or an ancestor.
func TestContextForFile(t *testing.T) {
	dir := seedContextFixture(t)
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"context", "--root", dir, "--file", "src/foo/bar/baz.go", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("context returned error: %v\n%s", err, out.String())
	}
	var recs []*record.Record
	if err := json.Unmarshal(out.Bytes(), &recs); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out.String())
	}
	if len(recs) != 3 {
		t.Fatalf("expected 3 records (file + 2 ancestors), got %d: %s", len(recs), out.String())
	}
	gotScopes := map[string]bool{}
	for _, r := range recs {
		gotScopes[r.Scope] = true
	}
	for _, want := range []string{"src/foo/bar/baz.go", "src/foo/bar", "src/foo"} {
		if !gotScopes[want] {
			t.Errorf("expected %q in scopes, got %v", want, gotScopes)
		}
	}
	for _, exclude := range []string{"src/other", "auth"} {
		if gotScopes[exclude] {
			t.Errorf("expected %q to be excluded, got %v", exclude, gotScopes)
		}
	}
}

// TestContextRadius confirms --radius caps the ancestor walk.
func TestContextRadius(t *testing.T) {
	dir := seedContextFixture(t)
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"context", "--root", dir, "--file", "src/foo/bar/baz.go", "--radius", "1", "--json"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("context returned error: %v", err)
	}
	var recs []*record.Record
	if err := json.Unmarshal(out.Bytes(), &recs); err != nil {
		t.Fatalf("output is not a JSON array of records: %v\n%s", err, out.String())
	}
	if len(recs) != 2 {
		t.Errorf("expected 2 records (file + parent) with radius=1, got %d", len(recs))
	}
	for _, r := range recs {
		if r.Scope != "src/foo/bar/baz.go" && r.Scope != "src/foo/bar" {
			t.Errorf("unexpected scope in radius=1 result: %q", r.Scope)
		}
	}
}

// TestContextMissingFile confirms --file is required.
func TestContextMissingFile(t *testing.T) {
	dir := seedContextFixture(t)
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"context", "--root", dir})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --file, got nil")
	}
	if !strings.Contains(err.Error(), "file") {
		t.Errorf("error should mention file, got: %v", err)
	}
}
